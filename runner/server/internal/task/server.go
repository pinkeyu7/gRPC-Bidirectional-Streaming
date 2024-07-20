package task

import (
	taskProto "grpc-bidirectional-streaming/pb/task"
	"grpc-bidirectional-streaming/pkg/helper"
	"grpc-bidirectional-streaming/pkg/prometheus"
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
	prometheus.TaskIdWorkerNum.Set(float64(helper.SyncMapLength(&s.taskIdWorkerMap)))
	prometheus.InputChanNum.Set(float64(helper.SyncMapLength(&s.inputChanMap)))
	prometheus.OutputChanNum.Set(float64(helper.SyncMapLength(&s.outputChanMap)))
}
