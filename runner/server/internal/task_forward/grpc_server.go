package task_forward

import (
	taskForwardProto "grpc-bidirectional-streaming/pb/task_forward"
	"grpc-bidirectional-streaming/pkg/grpc_streaming"
	"grpc-bidirectional-streaming/pkg/helper"
	"grpc-bidirectional-streaming/pkg/prometheus"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type Server struct {
	taskForwardProto.UnimplementedTaskForwardServer
	RequestChanMap  cmap.ConcurrentMap[string, *chan any]
	ResponseChanMap cmap.ConcurrentMap[string, *chan any]
}

func NewServer() *Server {
	return &Server{
		RequestChanMap:  cmap.New[*chan any](),
		ResponseChanMap: cmap.New[*chan any](),
	}
}

func (s *Server) Monitor() {
	prometheus.RequestChanNum.Set(float64(s.RequestChanMap.Count()))
	prometheus.ResponseChanNum.Set(float64(s.ResponseChanMap.Count()))
}

func (s *Server) Foo(stream taskForwardProto.TaskForward_FooServer) error {
	// Arrange
	tss := grpc_streaming.NewStreamingServer(helper.GetCurrentFunctionName(), stream, &s.RequestChanMap, &s.ResponseChanMap)

	// Read metadata from client
	err := tss.SetClientId(stream.Context())
	if err != nil {
		return status.Errorf(codes.DataLoss, err.Error())
	}

	// Setting streaming service
	err = tss.HandleStream()
	if err != nil {
		return status.Errorf(codes.Internal, "streaming failed: %v", err)
	}

	return nil
}
