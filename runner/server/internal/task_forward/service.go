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

func (s *Service) Unary(ctx context.Context, req *dto.UnaryRequest) (*dto.UnaryResponse, error) {
	// Jaeger
	ctx, span := jaeger.Tracer().Start(ctx, "request forward")
	span.AddEvent("init")
	defer span.End()

	res, err := grpc_streaming.ForwardUnaryRequestHandler[
		dto.UnaryRequest,
		dto.UnaryResponse,
		taskForwardProto.UnaryRequest,
		taskForwardProto.UnaryResponse,
	](ctx, s.mappingService, req.WorkerId, req)
	if err != nil {
		log.Printf("received error - worker id: %s, task id: %s, error: %s", req.WorkerId, req.TaskId, err.Error())
		return nil, err
	}

	span.AddEvent("done")

	return res, nil
}

func (s *Service) ClientStream(ctx context.Context, req *dto.ClientStreamRequest, resChan *chan *dto.ClientStreamResponse, errChan *chan error) {
	// Jaeger
	ctx, span := jaeger.Tracer().Start(ctx, "request forward")
	span.AddEvent("init")
	defer span.End()

	grpc_streaming.ForwardClientStreamRequestHandler[
		dto.ClientStreamRequest,
		dto.ClientStreamResponse,
		taskForwardProto.ClientStreamRequest,
		taskForwardProto.ClientStreamResponse,
	](ctx, s.mappingService, req.WorkerId, req, resChan, errChan)

	log.Printf("context done - service")
	span.AddEvent("context done - service")
}
