package grpcstreaming

import (
	"context"
	"fmt"
	"io"
	"log"

	"google.golang.org/grpc/metadata"
)

type serverObject[Request any, Response any] interface {
	Send(*Request) error
	Recv() (*Response, error)
	Context() context.Context
}

type server[Request any, Response any, Stream serverObject[Request, Response]] struct {
	clientID       string
	funcName       string
	mappingService *MappingService
	stream         serverObject[Request, Response]
}

func NewServer[Request any, Response any, Stream serverObject[Request, Response]](ms *MappingService, stream Stream) error {
	s := &server[Request, Response, Stream]{
		funcName:       getParentFunctionName(2),
		mappingService: ms,
		stream:         stream,
	}

	err := s.setClientID(stream.Context())
	if err != nil {
		return err
	}

	err = s.handleStream()
	if err != nil {
		return err
	}

	return nil
}

func (s *server[Request, Response, Stream]) setClientID(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return fmt.Errorf("could not extract grpc metadata")
	}
	if rs, ok := md["client_id"]; ok {
		for _, cid := range rs {
			s.clientID = cid
			return nil
		}
	}

	return fmt.Errorf("could not extract grpc client id")
}

func (s *server[Request, Response, Stream]) handleStream() error {
	// Get caller package name
	var reqStruct Request
	packageName := getPackageNameFromStruct(reqStruct)

	// Arrange
	requestChan := make(chan any)
	defer close(requestChan)

	s.mappingService.SetRequestChan(packageName, s.funcName, s.clientID, requestChan)

	// Request from client, send to worker
	go func() {
		for reqObj := range requestChan {
			req, ok := reqObj.(*Request)
			if !ok {
				log.Printf("failed to convert request")
				continue
			}

			if err := s.stream.Send(req); err != nil {
				log.Printf("failed to send request: %s", err.Error())
			}
		}
	}()

	// Receive from worker, send to client
	for {
		res, err := s.stream.Recv()
		if err == io.EOF {
			s.mappingService.RemoveRequestChan(packageName, s.funcName, s.clientID)
			return nil
		}
		if err != nil {
			s.mappingService.RemoveRequestChan(packageName, s.funcName, s.clientID)
			return err
		}

		go func(res *Response) {
			// Get requestId
			requestID, err := getFieldValue(res, "RequestId")
			if err != nil {
				log.Printf("failed to print request id: %s", err.Error())
				return
			}

			// Send to output channel
			responseChan, err := s.mappingService.GetResponseChan(requestID)
			if err != nil {
				log.Printf("request id: %s, error: %s", requestID, err.Error())
				return
			}

			// Send to output channel
			responseChan <- res
		}(res)
	}
}
