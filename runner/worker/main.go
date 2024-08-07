package main

import (
	"flag"
	"fmt"
	"grpc-bidirectional-streaming/config"
	taskProto "grpc-bidirectional-streaming/pb/task"
	grpc_streaming "grpc-bidirectional-streaming/pkg/grpc-streaming"
	"grpc-bidirectional-streaming/pkg/prometheus"
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
	tc := taskProto.NewTaskClient(conn)

	// Act
	go grpc_streaming.WorkerHandleStream[taskProto.RequestFromServerRequest, taskProto.RequestFromServerResponse, taskProto.Task_RequestFromServerClient](workerId, tc.RequestFromServer, ts.HandleRequest)

	select {}
}
