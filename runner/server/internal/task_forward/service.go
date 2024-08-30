package task_forward

import (
	"context"
	"fmt"
	"grpc-bidirectional-streaming/dto"
	taskForwardProto "grpc-bidirectional-streaming/pb/task_forward"
	"grpc-bidirectional-streaming/pkg/grpc_streaming"
	"grpc-bidirectional-streaming/pkg/jaeger"
	"log"
	"time"
)

type Service struct {
	mappingService *grpc_streaming.MappingService
}

func NewService(ms *grpc_streaming.MappingService) *Service {
	return &Service{
		mappingService: ms,
	}
}

func (s *Service) Foo(ctx context.Context, req *dto.FooRequest) (*dto.FooResponse, error) {
	// Jaeger
	ctx, span := jaeger.Tracer().Start(ctx, "request forward")
	span.AddEvent("init")
	defer span.End()

	res, err := grpc_streaming.ForwardRequestHandler[
		dto.FooRequest,
		dto.FooResponse,
		taskForwardProto.FooRequest,
		taskForwardProto.FooResponse,
	](ctx, s.mappingService, req.WorkerId, req)
	if err != nil {
		log.Printf("received error - worker id: %s, task id: %s, error: %s", req.WorkerId, req.TaskId, err.Error())
		return nil, err
	}

	span.AddEvent("done")

	return res, nil
}

func (s *Service) UpnpSearch(ctx context.Context, req *dto.UpnpSearchRequest, responseChan *chan *dto.UpnpSearchReply) {
	// Mock upnp search result
	resultChan := make(chan *dto.UpnpSearchReply)
	defer close(resultChan)

	go func() {
		for result := range resultChan {
			log.Printf("Received upnp search result: %+v", result)
			select {
			case <-ctx.Done():
				log.Println("context done")
				return
			default:
				*responseChan <- result
			}
		}
	}()

	for i := 0; i < 30; i++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("error: %v", r)
				}
			}()

			res := &dto.UpnpSearchReply{
				WorkerId: req.WorkerId,
				Model:    fmt.Sprintf("upnp_search_%d", i),
				Ip:       fmt.Sprintf("192.168.1.%d", i),
			}

			resultChan <- res
		}()
		time.Sleep(1 * time.Second)
	}
}
