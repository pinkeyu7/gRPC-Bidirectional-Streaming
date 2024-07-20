package task

import (
	"context"
	taskProto "grpc-bidirectional-streaming/pb/task"
	"log"
)

type Server struct {
	taskProto.UnimplementedTaskServer
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) RequestFromClient(context context.Context, req *taskProto.RequestFromClientRequest) (*taskProto.RequestFromClientResponse, error) {
	// Act
	log.Printf("Received: %v", req.GetTaskId())

	// Return
	// TODO ITEM - connect with worker chan
	res := &taskProto.RequestFromClientResponse{
		TaskId:      req.GetTaskId(),
		TaskMessage: "Hello World",
	}

	return res, nil
}
