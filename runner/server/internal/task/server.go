package task

import (
	taskProto "grpc-bidirectional-streaming/pb/task"
	"grpc-bidirectional-streaming/pkg/helper"
	"log"
	"sync"
)

type Server struct {
	taskProto.UnimplementedTaskServer
	taskIdWorkerMap sync.Map
	inputChanMap    sync.Map
	outputChanMap   sync.Map
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Monitor() {
	taskIdWorkerCount := helper.SyncMapLength(&s.taskIdWorkerMap)
	inputChanCount := helper.SyncMapLength(&s.inputChanMap)
	outputChanCount := helper.SyncMapLength(&s.outputChanMap)

	log.Printf("Monitor taskIdWorkerCount: %d, inputChanCount: %d, outputChanCount: %d", taskIdWorkerCount, inputChanCount, outputChanCount)
}
