package taskforward

import (
	taskForwardProto "grpc-bidirectional-streaming/pb/task_forward"
	grpcStreaming "grpc-bidirectional-streaming/pkg/grpc_streaming"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	taskForwardProto.UnimplementedTaskForwardServer
	mappingService *grpcStreaming.MappingService
}

func NewServer(ms *grpcStreaming.MappingService) *Server {
	return &Server{
		mappingService: ms,
	}
}

func (s *Server) Unary(stream taskForwardProto.TaskForward_UnaryServer) error {
	// Arrange
	err := grpcStreaming.NewServer(s.mappingService, stream)
	if err != nil {
		return status.Errorf(codes.Internal, "streaming failed: %s", err.Error())
	}

	return nil
}

func (s *Server) ClientStream(stream taskForwardProto.TaskForward_ClientStreamServer) error {
	// Arrange
	err := grpcStreaming.NewServer(s.mappingService, stream)
	if err != nil {
		return status.Errorf(codes.Internal, "streaming failed: %s", err.Error())
	}

	return nil
}
