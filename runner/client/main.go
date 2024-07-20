package main

import (
	"fmt"
	"grpc-bidirectional-streaming/config"
	"grpc-bidirectional-streaming/runner/client/internal/task"
	"log"
	"math/rand/v2"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	log.SetPrefix("[Client]")

	// Generate connection
	conn, err := grpc.NewClient(config.GetListenAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
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

			go func() {
				taskId := fmt.Sprintf("task_%s_%04d", "www", rand.IntN(config.GetTaskPerWorker())+1)
				err := taskClient.GetInfo(taskId)
				if err != nil {
					log.Printf("error: %v", err)
				}
				defer wg.Done()
			}()
		}
		time.Sleep(1 * time.Second)
	}

	// Wait for tasks
	wg.Wait()

	log.Println("done")
}
