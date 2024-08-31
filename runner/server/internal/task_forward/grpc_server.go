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

func (s *Server) Unary(stream taskForwardProto.TaskForward_UnaryServer) error {
	// Arrange
	err := grpc_streaming.NewServer(s.mappingService, stream)
	if err != nil {
		return status.Errorf(codes.Internal, "streaming failed: %s", err.Error())
	}

	return nil
}

func (s *Server) UpnpSearch(stream taskForwardProto.TaskForward_UpnpSearchServer) error {
	// Arrange
	err := grpc_streaming.NewServer(s.mappingService, stream)
	if err != nil {
		return status.Errorf(codes.Internal, "streaming failed: %s", err.Error())
	}

	return nil
}
