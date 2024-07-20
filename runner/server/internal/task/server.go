package task

import (
	taskProto "grpc-bidirectional-streaming/pb/task"
)

type Server struct {
	taskProto.UnimplementedTaskServer
	inputChan  chan *taskProto.RequestFromServerRequest
	outputChan chan *taskProto.RequestFromServerResponse
}

func NewServer() *Server {
	return &Server{
		inputChan:  make(chan *taskProto.RequestFromServerRequest),
		outputChan: make(chan *taskProto.RequestFromServerResponse),
	}
}
