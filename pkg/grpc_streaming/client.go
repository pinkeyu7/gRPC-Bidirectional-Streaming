package grpc_streaming

import (
	"context"
	"grpc-bidirectional-streaming/pkg/prometheus"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var clientId string

func SetClientId(cId string) {
	clientId = cId
}

type unaryClientObject[Request any, Response any] interface {
	Send(*Response) error
	Recv() (*Request, error)
}

type unaryClient[Request any, Response any, Client unaryClientObject[Request, Response]] struct {
	timeout   time.Duration
	handler   func(req *Request, resChan *chan *Response)
	getStream func(ctx context.Context, opts ...grpc.CallOption) (Client, error)
}

func NewUnaryClient[Request any, Response any, Client unaryClientObject[Request, Response]](
	ctx context.Context,
	getStream func(ctx context.Context, opts ...grpc.CallOption) (Client, error),
	handleRequest func(req *Request, resChan *chan *Response),
	timeout time.Duration,
) {

	c := &unaryClient[Request, Response, Client]{
		timeout:   timeout,
		handler:   handleRequest,
		getStream: getStream,
	}

	go c.handleUnary(ctx)
}

func (c *unaryClient[Request, Response, Client]) handleUnary(ctx context.Context) {
	// Arrange
	responseChan := make(chan *Response)
	defer close(responseChan)

	// Create metadata and context
	md := metadata.Pairs("client_id", clientId)
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

	// Handle message
	for {
		// Receive message
		req, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Printf("failed to receive request: %s", err.Error())
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
			subCtx, cancel := context.WithTimeout(context.Background(), c.timeout*time.Second)
			defer cancel()

			// Act
			resultChan := make(chan *Response, 1)
			defer close(resultChan)

			go c.handler(req, &resultChan)

			// Handle result
			select {
			case res := <-resultChan:
				responseChan <- res
			case <-subCtx.Done():
				requestId, err := getFieldValue(req, "RequestId")
				if err != nil {
					log.Printf("failed to get request id: %s", err.Error())
					return
				}
				responseChan <- CreateErrorResponse[Response](requestId, 0, "client - request timeout")
			}
		}(req)
	}
}
