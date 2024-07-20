package main

import (
	"grpc-bidirectional-streaming/config"
	"grpc-bidirectional-streaming/runner/worker/internal/task"
	taskService "grpc-bidirectional-streaming/runner/worker/internal/task/service"
	"log"

	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	log.SetPrefix("[Worker]")

	// Generate connection
	conn, err := grpc.NewClient(config.GetListenAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Init task service
	ts := taskService.NewService()
	taskClient := task.NewClient(conn, ts)

	// Act
	taskClient.GetInfo()
}
