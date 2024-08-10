package task_forward

import (
	"context"
	"grpc-bidirectional-streaming/dto"
	taskForwardProto "grpc-bidirectional-streaming/pb/task_forward"
	"grpc-bidirectional-streaming/pkg/grpc_streaming"
	"grpc-bidirectional-streaming/pkg/helper"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	server *Server
}

func NewService(s *Server) *Service {
	return &Service{
		server: s,
	}
}

func (s *Service) Foo(context context.Context, req *dto.FooRequest) (*dto.FooResponse, error) {
	// Arrange
	reqTo := &taskForwardProto.FooRequest{
		TaskId: req.TaskId,
	}

	resObj, err := grpc_streaming.HandleRequest(context, req.WorkerId, helper.GetCurrentFunctionName(), reqTo, &s.server.RequestChanMap, &s.server.ResponseChanMap)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	res, ok := resObj.(*taskForwardProto.FooResponse)
	if !ok {
		return nil, status.Errorf(codes.Internal, "response convert error")
	}

	return &dto.FooResponse{
		TaskId:      res.GetTaskId(),
		TaskMessage: res.GetTaskMessage(),
	}, nil
}
