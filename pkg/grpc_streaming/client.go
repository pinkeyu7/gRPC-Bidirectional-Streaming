package grpc_streaming

import (
	"context"
	"grpc-bidirectional-streaming/pkg/prometheus"
	"io"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type StreamingClientObject[Request any, Response any] interface {
	Send(*Response) error
	Recv() (*Request, error)
}

type StreamingClient[Request any, Response any, Client StreamingClientObject[Request, Response]] struct {
	clientId  string
	handler   func(req *Request) (*Response, error)
	getStream func(ctx context.Context, opts ...grpc.CallOption) (Client, error)
	doneChan  chan bool
}

func NewStreamingClient[Request any, Response any, Client StreamingClientObject[Request, Response]](clientId string, getStream func(ctx context.Context, opts ...grpc.CallOption) (Client, error), handleRequest func(req *Request) (*Response, error)) *StreamingClient[Request, Response, Client] {
	return &StreamingClient[Request, Response, Client]{
		clientId:  clientId,
		handler:   handleRequest,
		getStream: getStream,
		doneChan:  make(chan bool, 1),
	}
}

func (c *StreamingClient[Request, Response, Client]) HandleStream(context context.Context) {
	// Arrange
	responseChan := make(chan *Response)
	defer close(responseChan)

	// Create metadata and context
	md := metadata.Pairs("client_id", c.clientId)
	ctx := metadata.NewOutgoingContext(context, md)

	// Make RPC using the context with the metadata
	stream, err := c.getStream(ctx)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// Send to server
	go func() {
		for req := range responseChan {
			if err := stream.Send(req); err != nil {
				log.Printf("failed to send request: %v", err)
			}
			prometheus.ResponseNum.Inc()
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
			log.Printf("failed to receive request: %v", err)
			return
		}

		prometheus.RequestNum.Inc()

		go func(req *Request) {
			res, err := c.handler(req)
			if err != nil {
				log.Printf("failed to handle request: %v", err)
			}

			responseChan <- res
		}(req)
	}
}

func (c *StreamingClient[Request, Response, Client]) Shutdown() {
	log.Printf("client id: %s, shutting down", c.clientId)
	c.doneChan <- true
}
