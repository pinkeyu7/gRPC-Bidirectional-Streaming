package main

import (
	"context"
	"grpc-bidirectional-streaming/config"
	taskProto "grpc-bidirectional-streaming/pb/task"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	log.SetPrefix("[Client]")

	conn, err := grpc.NewClient(config.GetListenAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := taskProto.NewTaskClient(conn)

	// Send request
	req := &taskProto.RequestFromClientRequest{
		TaskId: "test_task_id",
	}
	res, err := client.RequestFromClient(context.Background(), req)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	log.Printf("task id: %s, task message: %s", res.GetTaskId(), res.GetTaskMessage())
}
