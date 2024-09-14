package grpcstreaming

import (
	"context"
	connectionStatus "grpc-bidirectional-streaming/pkg/connection_status"
	"grpc-bidirectional-streaming/pkg/prometheus"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type clientStreamClient[Request any, Response any, Client clientObject[Request, Response]] struct {
	timeout          time.Duration
	getStream        func(ctx context.Context, opts ...grpc.CallOption) (Client, error)
	handler          func(ctx context.Context, req *Request, resChan chan *Response)
	connectionStatus *connectionStatus.ConnectionStatus
}

func NewClientStreamClient[Request any, Response any, Client clientObject[Request, Response]](
	ctx context.Context,
	getStream func(ctx context.Context, opts ...grpc.CallOption) (Client, error),
	handler func(ctx context.Context, req *Request, resChan chan *Response),
	timeout time.Duration,
) {

	c := &clientStreamClient[Request, Response, Client]{
		timeout:          timeout,
		getStream:        getStream,
		handler:          handler,
		connectionStatus: connectionStatus.NewConnectionStatus(),
	}

	// Start connection when received trigger event
	go func() {
		for range c.connectionStatus.EventChan() {
			go c.handleClientStream(ctx)
		}
	}()

	// Start service
	c.connectionStatus.Start()
}

func (c *clientStreamClient[Request, Response, Client]) handleClientStream(ctx context.Context) {
	log.Println("client stream client started")

	// Defer func
	defer func() {
		log.Println("client stream client end")
		c.connectionStatus.DeferClose()
	}()

	// Arrange
	responseChan := make(chan *Response)
	defer close(responseChan)

	// Create metadata and context
	md := metadata.Pairs("client_id", clientID)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Make RPC using the context with the metadata
	stream, err := c.getStream(ctx)
	if err != nil {
		log.Printf("error: %s", err.Error())
		return
	}

	// Send to server
	go func() {
		for req := range responseChan {
			if err := stream.Send(req); err != nil {
				log.Printf("failed to send request: %s", err.Error())
			}
		}
	}()

	// Send connected
	c.connectionStatus.Connected()

	// Handle message
	for {
		// Receive message
		req, err := stream.Recv()
		if err == io.EOF {
			log.Println("stream closed with EOF")
			return
		}
		if err != nil {
			log.Printf("failed to receive request: %s", err.Error())
			c.connectionStatus.Error(err)
			return
		}

		prometheus.RequestNum.Inc()

		go func(req *Request) {
			// Defer func to prevent sent to close channel
			defer func() {
				if r := recover(); r != nil {
					log.Println("recover from client closed the stream")
				}
			}()

			// Arrange sub context for time out handling
			subCtx, cancel := context.WithTimeout(context.Background(), c.timeout)
			defer cancel()

			// Arrange channel
			resultChan := make(chan *Response, 1)
			defer close(resultChan)

			// Act
			go c.handler(subCtx, req, resultChan)

			// Handle result
			for {
				select {
				case res := <-resultChan:
					responseChan <- res
				case <-subCtx.Done():
					requestID, err := getFieldValue(req, "RequestId")
					if err != nil {
						log.Printf("failed to get request id: %s", err.Error())
						return
					}
					responseChan <- NewErrorResponse[Response](requestID, ErrorCodeClientTimeout, "client - request timeout")
					return
				}
			}
		}(req)
	}
}
