package task_forward

import (
	taskForwardProto "grpc-bidirectional-streaming/pb/task_forward"
	"grpc-bidirectional-streaming/pkg/grpc_streaming"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	taskForwardProto.UnimplementedTaskForwardServer
	mappingService *grpc_streaming.MappingService
}

func NewServer(ms *grpc_streaming.MappingService) *Server {
	return &Server{
		mappingService: ms,
	}
}

func (s *Server) Foo(stream taskForwardProto.TaskForward_FooServer) error {
	// Arrange
	err := grpc_streaming.NewStreamingServer(s.mappingService, stream)
	if err != nil {
		return status.Errorf(codes.Internal, "streaming failed: %v", err)
	}

	return nil
}
