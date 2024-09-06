package main

import (
	"context"
	"flag"
	"fmt"
	"grpc-bidirectional-streaming/config"
	taskForwardProto "grpc-bidirectional-streaming/pb/task_forward"
	grpcStreaming "grpc-bidirectional-streaming/pkg/grpc_streaming"
	"grpc-bidirectional-streaming/pkg/prometheus"
	taskForward "grpc-bidirectional-streaming/runner/worker/internal/task_forward"
	"log"
	"net"
	"os/signal"
	"syscall"

	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var workerID string

func main() {
	flag.StringVar(&workerID, "workerID", config.GetWorkerID(), "worker id")
	flag.Parse()

	log.SetPrefix(fmt.Sprintf("[Worker: %s]", workerID))

	// Pusher
	pusher := prometheus.NewPusher(fmt.Sprintf("worker_%s", workerID))
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
		log.Fatalf("did not connect: %s", err.Error())
	}
	defer conn.Close()

	// Arrange context
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Init task service
	tfs := taskForward.NewService()
	tfs.InitTaskMessage(workerID, config.GetTaskPerWorker())
	tfc := taskForwardProto.NewTaskForwardClient(conn)

	// Act
	grpcStreaming.SetClientID(workerID)
	grpcStreaming.NewUnaryClient(ctx, tfc.Unary, tfs.Unary, config.GetClientTimeout())
	grpcStreaming.NewClientStreamClient(ctx, tfc.ClientStream, tfs.ClientStream, config.GetClientTimeout())

	// Graceful shutdown
	<-ctx.Done()
	log.Println("worker shutting down")
}
