package task

import (
	taskProto "grpc-bidirectional-streaming/pb/task"
	"grpc-bidirectional-streaming/pkg/prometheus"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type Server struct {
	taskProto.UnimplementedTaskServer
	inputChanMap  cmap.ConcurrentMap[string, *chan *taskProto.RequestFromServerRequest]
	outputChanMap cmap.ConcurrentMap[string, *cmap.ConcurrentMap[string, *chan *taskProto.RequestFromServerResponse]]
}

func NewServer() *Server {
	return &Server{
		inputChanMap:  cmap.New[*chan *taskProto.RequestFromServerRequest](),
		outputChanMap: cmap.New[*cmap.ConcurrentMap[string, *chan *taskProto.RequestFromServerResponse]](),
	}
}

func (s *Server) Monitor() {
	prometheus.InputChanNum.Set(float64(s.inputChanMap.Count()))
	prometheus.OutputChanNum.Set(float64(s.outputChanMap.Count()))
}
