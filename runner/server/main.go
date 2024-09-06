package main

import (
	"context"
	"grpc-bidirectional-streaming/config"
	taskProto "grpc-bidirectional-streaming/pb/task"
	taskForwardProto "grpc-bidirectional-streaming/pb/task_forward"
	grpcStreaming "grpc-bidirectional-streaming/pkg/grpc_streaming"
	"grpc-bidirectional-streaming/pkg/jaeger"
	"grpc-bidirectional-streaming/pkg/prometheus"
	"grpc-bidirectional-streaming/runner/server/internal/task"
	taskForward "grpc-bidirectional-streaming/runner/server/internal/task_forward"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
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
			log.Printf("Error shutting down tracer provider: %s", err.Error())
		}
	}()

	// Pusher
	pusher := prometheus.NewPusher("server")
	pusher.Start()

	// Init
	ms := grpcStreaming.NewMappingService()
	tfs := taskForward.NewService(ms)
	tfgs := taskForward.NewServer(ms)
	tgs := task.NewServer(tfs)

	go func() {
		for {
			ms.Monitor()
			time.Sleep(config.GetMonitorTimeInterval())
		}
	}()

	// Listen
	lis, err := net.Listen(config.GetListenNetwork(), config.GetListenAddress())
	if err != nil {
		log.Printf("failed to listen: %s", err.Error())
		panic(err)
	}

	// Start Server
	s := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	taskForwardProto.RegisterTaskForwardServer(s, tfgs)
	taskProto.RegisterTaskServer(s, tgs)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Printf("failed to serve: %s", err.Error())
		panic(err)
	}
}
