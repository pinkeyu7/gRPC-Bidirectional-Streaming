package main

import (
	"context"
	"grpc-bidirectional-streaming/config"
	taskProto "grpc-bidirectional-streaming/pb/task"
	taskForwardProto "grpc-bidirectional-streaming/pb/task_forward"
	"grpc-bidirectional-streaming/pkg/grpc_streaming"
	"grpc-bidirectional-streaming/pkg/jaeger"
	"grpc-bidirectional-streaming/pkg/prometheus"
	"grpc-bidirectional-streaming/runner/server/internal/task"
	"grpc-bidirectional-streaming/runner/server/internal/task_forward"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	_ "github.com/joho/godotenv/autoload"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

func main() {
	if config.GetListenNetwork() == "unix" {
		_ = os.Remove(config.GetListenAddress())
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		if config.GetListenNetwork() == "unix" {
			_ = os.Remove(config.GetListenAddress())
		}

		os.Exit(0)
	}()

	log.SetPrefix("[Server]")

	// Jaeger
	tp, err := jaeger.InitTracer(context.Background(), "server")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// Pusher
	pusher := prometheus.NewPusher("server")
	pusher.Start()

	// Init
	ms := grpc_streaming.NewMappingService()
	tfs := task_forward.NewService(ms)
	tfgs := task_forward.NewServer(ms)
	tgs := task.NewServer(tfs)

	go func() {
		for {
			ms.Monitor()
			time.Sleep(5 * time.Second)
		}
	}()

	// Listen
	lis, err := net.Listen(config.GetListenNetwork(), config.GetListenAddress())
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Start Server
	s := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	taskForwardProto.RegisterTaskForwardServer(s, tfgs)
	taskProto.RegisterTaskServer(s, tgs)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
