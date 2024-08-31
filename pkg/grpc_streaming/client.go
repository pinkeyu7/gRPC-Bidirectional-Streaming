package grpc_streaming

import (
	"context"
	"grpc-bidirectional-streaming/pkg/prometheus"
	"io"
	"log"

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
	handler   func(req *Request) (*Response, error)
	getStream func(ctx context.Context, opts ...grpc.CallOption) (Client, error)
}

func NewUnaryClient[Request any, Response any, Client unaryClientObject[Request, Response]](context context.Context, getStream func(ctx context.Context, opts ...grpc.CallOption) (Client, error), handleRequest func(req *Request) (*Response, error)) {
	c := &unaryClient[Request, Response, Client]{
		handler:   handleRequest,
		getStream: getStream,
	}

	go c.handleUnary(context)
}

func (c *unaryClient[Request, Response, Client]) handleUnary(context context.Context) {
	// Arrange
	responseChan := make(chan *Response)
	defer close(responseChan)

	// Create metadata and context
	md := metadata.Pairs("client_id", clientId)
	ctx := metadata.NewOutgoingContext(context, md)

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
			defer func() {
				if r := recover(); r != nil {
					log.Println("recover from client closed the stream")
				}
			}()

			res, err := c.handler(req)
			if err != nil {
				log.Printf("failed to handle request: %s", err.Error())
			}

			responseChan <- res
		}(req)
	}
}
