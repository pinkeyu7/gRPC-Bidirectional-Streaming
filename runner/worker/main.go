package main

import (
	"flag"
	"fmt"
	"grpc-bidirectional-streaming/config"
	"grpc-bidirectional-streaming/pkg/prometheus"
	"grpc-bidirectional-streaming/runner/worker/internal/task"
	taskService "grpc-bidirectional-streaming/runner/worker/internal/task/service"
	"log"
	"net"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"context"
)

var workerId string

func main() {
	time.Sleep(5 * time.Second)

	flag.StringVar(&workerId, "workerId", config.GetWorkerId(), "worker id")
	flag.Parse()

	log.SetPrefix(fmt.Sprintf("[Worker: %s]", workerId))

	// Pusher
	pusher := prometheus.NewPusher(fmt.Sprintf("worker_%s", workerId))
	pusher.Start()

	// Generate connection
	conn, err := grpc.NewClient(fmt.Sprintf("passthrough:%s", config.GetListenAddress()),
		grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(
			ctx context.Context, s string,
		) (net.Conn, error) {
			log.Printf("Dialing %s\n", config.GetListenAddress())
			return net.Dial(config.GetListenNetwork(), config.GetListenAddress())
		}))
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
