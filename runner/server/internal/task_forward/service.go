package task_forward

import (
	"context"
	"grpc-bidirectional-streaming/dto"
	taskForwardProto "grpc-bidirectional-streaming/pb/task_forward"
	"grpc-bidirectional-streaming/pkg/grpc_streaming"
	"grpc-bidirectional-streaming/pkg/jaeger"
	"log"
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

	res, err := grpc_streaming.ForwardUnaryRequestHandler[
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

func (s *Service) UpnpSearch(ctx context.Context, req *dto.UpnpSearchRequest, responseChan *chan *dto.UpnpSearchResponse, errChan *chan error) {
	// Jaeger
	ctx, span := jaeger.Tracer().Start(ctx, "request forward")
	span.AddEvent("init")
	defer span.End()

	grpc_streaming.ForwardClientStreamRequestHandler[
		dto.UpnpSearchRequest,
		dto.UpnpSearchResponse,
		taskForwardProto.UpnpSearchRequest,
		taskForwardProto.UpnpSearchResponse,
	](ctx, s.mappingService, req.WorkerId, req, responseChan, errChan)

	log.Printf("context done - service")
	span.AddEvent("context done - service")
}
