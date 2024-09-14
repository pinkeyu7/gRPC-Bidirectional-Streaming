package grpcstreaming

import (
	"context"
	"grpc-bidirectional-streaming/pkg/prometheus"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type unaryClient[Request any, Response any, Client clientObject[Request, Response]] struct {
	timeout         time.Duration
	getStream       func(ctx context.Context, opts ...grpc.CallOption) (Client, error)
	handler         func(ctx context.Context, req *Request, resChan chan *Response)
	connectionAgent *connectionAgent
}

func NewUnaryClient[Request any, Response any, Client clientObject[Request, Response]](
	ctx context.Context,
	getStream func(ctx context.Context, opts ...grpc.CallOption) (Client, error),
	handler func(ctx context.Context, req *Request, resChan chan *Response),
	timeout time.Duration,
) {

	c := &unaryClient[Request, Response, Client]{
		timeout:   timeout,
		getStream: getStream,
		handler:   handler,
	}

	// Set connection handler
	c.connectionAgent = newConnectionAgent(ctx, c.handleUnary)
	c.connectionAgent.Start()
}

func (c *unaryClient[Request, Response, Client]) handleUnary(ctx context.Context) {
	log.Println("unary client start")

	// Defer func
	defer func() {
		log.Println("unary client end")
		c.connectionAgent.Disconnected()
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
	c.connectionAgent.Connected()

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
			c.connectionAgent.Error(err)
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

			// Act
			resultChan := make(chan *Response, 1)
			defer close(resultChan)

			go c.handler(subCtx, req, resultChan)

			// Handle result
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
			}
		}(req)
	}
}
