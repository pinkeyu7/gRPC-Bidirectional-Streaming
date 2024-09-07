package taskforward

import (
	"context"
	"grpc-bidirectional-streaming/dto"
	taskForwardProto "grpc-bidirectional-streaming/pb/task_forward"
	grpcStreaming "grpc-bidirectional-streaming/pkg/grpc_streaming"
	"grpc-bidirectional-streaming/pkg/jaeger"
	"log"
)

type Service struct {
	mappingService *grpcStreaming.MappingService
}

func NewService(ms *grpcStreaming.MappingService) *Service {
	return &Service{
		mappingService: ms,
	}
}

func (s *Service) Unary(ctx context.Context, req *dto.UnaryRequest) (*dto.UnaryResponse, *grpcStreaming.ErrorInfo) {
	// Jaeger
	ctx, span := jaeger.Tracer().Start(ctx, "request forward")
	span.AddEvent("init")
	defer span.End()

	res, err := grpcStreaming.ForwardUnaryRequestHandler[
		dto.UnaryRequest,
		dto.UnaryResponse,
		taskForwardProto.UnaryRequest,
		taskForwardProto.UnaryResponse,
	](ctx, s.mappingService, req.WorkerID, req)
	if err != nil {
		log.Printf("received error - worker id: %s, task id: %s, error: %s", req.WorkerID, req.TaskID, err.Message)
		return nil, err
	}

	span.AddEvent("done")

	return res, nil
}

func (s *Service) ClientStream(ctx context.Context, req *dto.ClientStreamRequest, resChan chan *dto.ClientStreamResponse,
	errChan chan *grpcStreaming.ErrorInfo) {

	// Jaeger
	ctx, span := jaeger.Tracer().Start(ctx, "request forward")
	span.AddEvent("init")
	defer span.End()

	grpcStreaming.ForwardClientStreamRequestHandler[
		dto.ClientStreamRequest,
		dto.ClientStreamResponse,
		taskForwardProto.ClientStreamRequest,
		taskForwardProto.ClientStreamResponse,
	](ctx, s.mappingService, req.WorkerID, req, resChan, errChan)

	log.Printf("context done - service")
	span.AddEvent("context done - service")
}
