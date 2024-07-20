package main

import (
	"flag"
	"fmt"
	"grpc-bidirectional-streaming/config"
	"grpc-bidirectional-streaming/pkg/prometheus"
	"grpc-bidirectional-streaming/runner/worker/internal/task"
	taskService "grpc-bidirectional-streaming/runner/worker/internal/task/service"
	"log"

	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var workerId string

func main() {
	flag.StringVar(&workerId, "workerId", "worker_default", "worker id")
	flag.Parse()

	log.SetPrefix(fmt.Sprintf("[Worker: %s]", workerId))

	// Pusher
	pusher := prometheus.NewPusher(fmt.Sprintf("worker_%s", workerId))
	pusher.Start()

	// Generate connection
	conn, err := grpc.NewClient(config.GetListenAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Init task service
	ts := taskService.NewService(workerId, config.GetTaskPerWorker())
	taskClient := task.NewClient(workerId, conn, ts)

	// Register workerId
	err = taskClient.RegisterIds()
	if err != nil {
		log.Fatalf("could not register ids: %v", err)
	}

	// Act
	taskClient.GetInfo()
}
