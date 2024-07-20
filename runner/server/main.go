package main

import (
	"grpc-bidirectional-streaming/config"
	taskProto "grpc-bidirectional-streaming/pb/task"
	"grpc-bidirectional-streaming/runner/server/internal/task"
	"log"
	"net"

	"google.golang.org/grpc"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	log.SetPrefix("[Server]")

	// Init
	ts := task.NewServer()

	// Listen
	lis, err := net.Listen(config.GetListenNetwork(), config.GetListenAddress())
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Start Server
	s := grpc.NewServer()
	taskProto.RegisterTaskServer(s, ts)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
