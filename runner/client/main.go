package main

import (
	"context"
	"fmt"
	"grpc-bidirectional-streaming/config"
	"grpc-bidirectional-streaming/pkg/prometheus"
	"grpc-bidirectional-streaming/runner/client/internal/task"
	"log"
	"math/rand/v2"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	log.SetPrefix("[Client]")

	// Pusher
	pusher := prometheus.NewPusher("client")
	stopChan := pusher.Start()

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
	taskClient := task.NewClient(conn)

	// Act
	wg := new(sync.WaitGroup)

	for i := 0; i < config.GetRequestTimeDuration(); i++ {
		for j := 0; j < config.GetRequestPerSecond(); j++ {
			wg.Add(1)
			prometheus.RequestNum.Inc()

			go func() {
				start := time.Now()
				taskId := fmt.Sprintf("task_%s_%04d", fmt.Sprintf("worker_%03d", rand.IntN(config.GetWorkerCount())+1), rand.IntN(config.GetTaskPerWorker())+1)

				err := taskClient.GetInfo(taskId)
				duration := time.Since(start)
				if err != nil {
					log.Printf("error: %v", err)
					prometheus.ResponseTime.WithLabelValues("fail").Observe(duration.Seconds())
				} else {
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

	*stopChan <- true
	close(*stopChan)

	log.Println("done")
}
