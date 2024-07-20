package task

import taskProto "grpc-bidirectional-streaming/pb/task"

type Server struct {
	taskProto.UnimplementedTaskServer
}

func NewServer() *Server {
	return &Server{}
}
