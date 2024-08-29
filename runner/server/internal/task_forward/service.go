package task_forward

import (
	"context"
	"grpc-bidirectional-streaming/dto"
	taskForwardProto "grpc-bidirectional-streaming/pb/task_forward"
	"grpc-bidirectional-streaming/pkg/grpc_streaming"
	"grpc-bidirectional-streaming/pkg/jaeger"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	// Arrange
	reqTo := &taskForwardProto.FooRequest{
		TaskId: req.TaskId,
	}

	resObj, err := grpc_streaming.HandleRequest(ctx, s.mappingService, req.WorkerId, reqTo)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	res, ok := resObj.(*taskForwardProto.FooResponse)
	if !ok {
		return nil, status.Errorf(codes.Internal, "response convert error")
	}

	span.AddEvent("done")

	return &dto.FooResponse{
		WorkerId:    req.WorkerId,
		TaskId:      res.GetTaskId(),
		TaskMessage: res.GetTaskMessage(),
	}, nil
}
