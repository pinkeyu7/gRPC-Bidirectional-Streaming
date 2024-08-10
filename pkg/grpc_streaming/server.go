package grpc_streaming

import (
	"context"
	"fmt"
	"grpc-bidirectional-streaming/pkg/helper"
	"io"
	"log"

	cmap "github.com/orcaman/concurrent-map/v2"

	"google.golang.org/grpc/metadata"
)

type StreamingServerObject[Request any, Response any] interface {
	Send(*Request) error
	Recv() (*Response, error)
}

type StreamingServer[Request any, Response any, Stream StreamingServerObject[Request, Response]] struct {
	clientId        string
	funcName        string
	stream          StreamingServerObject[Request, Response]
	requestChanMap  *cmap.ConcurrentMap[string, *chan any]
	responseChanMap *cmap.ConcurrentMap[string, *chan any]
	doneChan        chan bool
}

func NewStreamingServer[Request any, Response any, Stream StreamingServerObject[Request, Response]](funcName string, stream Stream, requestChanMap *cmap.ConcurrentMap[string, *chan any], responseChanMap *cmap.ConcurrentMap[string, *chan any]) *StreamingServer[Request, Response, Stream] {
	return &StreamingServer[Request, Response, Stream]{
		funcName:        funcName,
		stream:          stream,
		requestChanMap:  requestChanMap,
		responseChanMap: responseChanMap,
		doneChan:        make(chan bool, 1),
	}
}

func (s *StreamingServer[Request, Response, Stream]) SetClientId(context context.Context) error {
	md, ok := metadata.FromIncomingContext(context)
	if !ok {
		return fmt.Errorf("could not extract grpc metadata")
	}
	if rs, ok := md["client_id"]; ok {
		for _, cid := range rs {
			s.clientId = cid
			return nil
		}
	}

	return fmt.Errorf("could not extract grpc client id")
}

func (s *StreamingServer[Request, Response, Stream]) HandleStream() error {
	// Arrange
	requestChan := make(chan any)
	defer close(requestChan)

	requestChanIndex := GetChanIndex(s.clientId, s.funcName)
	s.requestChanMap.Set(requestChanIndex, &requestChan)

	// Request from client, send to worker
	go func() {
		for reqObj := range requestChan {
			req, ok := reqObj.(*Request)
			if !ok {
				log.Printf("failed to convert request")
				continue
			}

			if err := s.stream.Send(req); err != nil {
				log.Printf("failed to send request: %v", err)
			}
		}
	}()

	// Receive from worker, send to client
	for {
		res, err := s.stream.Recv()
		if err == io.EOF {
			s.requestChanMap.Remove(requestChanIndex)
			return nil
		}
		if err != nil {
			s.requestChanMap.Remove(requestChanIndex)
			return err
		}

		go func(res *Response) {
			// Get requestId
			requestId, err := helper.GetFieldValue(res, "RequestId")
			if err != nil {
				log.Printf("failed to print request id: %v", err)
				return
			}

			// Send to output channel
			responseChan, ok := s.responseChanMap.Get(requestId)
			if !ok {
				log.Printf("failed to find output channel: request id: %v", requestId)
				return
			}

			// Send to output channel
			*responseChan <- res
			s.responseChanMap.Remove(requestId)
		}(res)
	}
}

func GetChanIndex(clientId string, funcName string) string {
	return fmt.Sprintf("%s_%s", clientId, funcName)
}
