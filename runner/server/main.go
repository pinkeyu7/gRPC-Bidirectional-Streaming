package main

import (
	"context"
	"grpc-bidirectional-streaming/config"
	taskProto "grpc-bidirectional-streaming/pb/task"
	"grpc-bidirectional-streaming/pkg/jaeger"
	"grpc-bidirectional-streaming/pkg/prometheus"
	"grpc-bidirectional-streaming/runner/server/internal/task"
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
	ts := task.NewServer()

	go func() {
		for {
			ts.Monitor()
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
	taskProto.RegisterTaskServer(s, ts)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
