package main

import (
	"context"
	"fmt"
	"grpc-bidirectional-streaming/config"
	"grpc-bidirectional-streaming/pkg/helper"
	"grpc-bidirectional-streaming/pkg/jaeger"
	"grpc-bidirectional-streaming/pkg/prometheus"
	"grpc-bidirectional-streaming/runner/client/internal/task"
	"log"
	"net"
	"sync"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	log.SetPrefix("[Client]")

	// Jaeger
	tp, err := jaeger.InitTracer(context.Background(), "client")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %s", err.Error())
		}
	}()

	// Pusher
	pusher := prometheus.NewPusher("client")
	stopChan := pusher.Start()

	// Generate connection
	conn, err := grpc.NewClient(fmt.Sprintf("passthrough:%s", config.GetListenAddress()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithContextDialer(func(
			ctx context.Context, s string,
		) (net.Conn, error) {

			log.Printf("Dialing %s\n", config.GetListenAddress())
			return net.Dial(config.GetListenNetwork(), config.GetListenAddress())
		}))
	if err != nil {
		log.Printf("did not connect: %s", err.Error())
		panic(err)
	}
	defer conn.Close()

	// Init task service
	taskClient := task.NewClient(conn)
	taskClient.InitTaskMessage(config.GetWorkerID(), config.GetTaskPerWorker())

	successNum := 0
	failNum := 0

	// Act
	wg := new(sync.WaitGroup)

	for i := 0; i < config.GetRequestTimeDuration(); i++ {
		for j := 0; j < config.GetRequestPerSecond(); j++ {
			wg.Add(1)
			prometheus.RequestNum.Inc()

			go func() {
				start := time.Now()
				workerID := fmt.Sprintf("worker_%03d", helper.RandomInt(0, config.GetWorkerCount())+1)
				taskID := fmt.Sprintf("task_%s_%04d", workerID, helper.RandomInt(0, config.GetTaskPerWorker())+1)

				err := taskClient.GetInfo(context.Background(), workerID, taskID)
				duration := time.Since(start)
				if err != nil {
					log.Printf("worker id: %s, task id: %s, error: %s", workerID, taskID, err.Error())
					failNum++
					prometheus.ResponseTime.WithLabelValues("fail").Observe(duration.Seconds())
				} else {
					successNum++
					prometheus.ResponseTime.WithLabelValues("success").Observe(duration.Seconds())
				}
				prometheus.ResponseNum.Inc()

				defer wg.Done()
			}()
		}
		time.Sleep(1 * time.Second)
	}

	// Wait for tasks
	wg.Wait()

	stopChan <- true
	close(stopChan)

	log.Printf("done: success: %d, fail: %d", successNum, failNum)
}
