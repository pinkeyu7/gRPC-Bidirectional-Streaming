package task

import (
	taskProto "grpc-bidirectional-streaming/pb/task"
	"grpc-bidirectional-streaming/pkg/prometheus"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type Server struct {
	taskProto.UnimplementedTaskServer
	taskIdWorkerMap cmap.ConcurrentMap[string, string]
	inputChanMap    cmap.ConcurrentMap[string, *chan *taskProto.RequestFromServerRequest]
	outputChanMap   cmap.ConcurrentMap[string, *cmap.ConcurrentMap[string, *chan *taskProto.RequestFromServerResponse]]
}

func NewServer() *Server {
	return &Server{
		taskIdWorkerMap: cmap.New[string](),
		inputChanMap:    cmap.New[*chan *taskProto.RequestFromServerRequest](),
		outputChanMap:   cmap.New[*cmap.ConcurrentMap[string, *chan *taskProto.RequestFromServerResponse]](),
	}
}

func (s *Server) Monitor() {
	prometheus.TaskIdWorkerNum.Set(float64(s.taskIdWorkerMap.Count()))
	prometheus.InputChanNum.Set(float64(s.inputChanMap.Count()))
	prometheus.OutputChanNum.Set(float64(s.outputChanMap.Count()))
}
