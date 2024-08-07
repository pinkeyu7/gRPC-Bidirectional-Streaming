package grpc_streaming

import (
	"context"
	"grpc-bidirectional-streaming/pkg/prometheus"
	"io"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type StreamObj[Request any, Response any] interface {
	Send(*Response) error
	Recv() (*Request, error)
	grpc.ClientStream
}

type GetStream[Request any, Response any, Client StreamObj[Request, Response]] func(ctx context.Context, opts ...grpc.CallOption) (Client, error)
type HandleRequest[Request any, Response any] func(req *Request) (*Response, error)

func WorkerHandleStream[Request any, Response any, Client StreamObj[Request, Response]](workerId string, getStream GetStream[Request, Response, Client], handleRequest HandleRequest[Request, Response]) {
	// Arrange
	outputChan := make(chan *Response)
	defer close(outputChan)

	// Create metadata and context
	md := metadata.Pairs("worker_id", workerId)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// Make RPC using the context with the metadata
	stream, err := getStream(ctx)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// Send to server
	go func() {
		for req := range outputChan {
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
			log.Fatalf("failed to receive request: %v", err)
		}

		prometheus.RequestNum.Inc()

		go func(req *Request) {
			prometheus.WorkerRequestNum.Inc()

			res, err := handleRequest(req)
			if err != nil {
				log.Printf("failed to handle request: %v", err)
			}

			outputChan <- res
		}(req)
	}
}
