package task

import (
	"context"
	"grpc-bidirectional-streaming/dto"
	taskProto "grpc-bidirectional-streaming/pb/task"
	"grpc-bidirectional-streaming/runner/server/internal/task_forward"
)

type Server struct {
	taskProto.UnimplementedTaskServer
	taskForwardService *task_forward.Service
}

func NewServer(tfs *task_forward.Service) *Server {
	return &Server{taskForwardService: tfs}
}

func (s *Server) Foo(context context.Context, req *taskProto.FooRequest) (*taskProto.FooResponse, error) {
	request := &dto.FooRequest{
		WorkerId: req.GetWorkerId(),
		TaskId:   req.TaskId,
	}

	response, err := s.taskForwardService.Foo(context, request)
	if err != nil {
		return nil, err
	}

	return &taskProto.FooResponse{
		WorkerId:    response.WorkerId,
		TaskId:      response.TaskId,
		TaskMessage: response.TaskMessage,
	}, nil
}
