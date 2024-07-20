package task

import (
	taskProto "grpc-bidirectional-streaming/pb/task"
	"sync"
)

type Server struct {
	taskProto.UnimplementedTaskServer
	inputChanMap  sync.Map
	outputChanMap sync.Map
}

func NewServer() *Server {
	return &Server{}
}
