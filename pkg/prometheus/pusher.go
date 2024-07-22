package prometheus

import (
	"fmt"
	"grpc-bidirectional-streaming/config"
	"log"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus/push"
)

type Pusher struct {
	jobName string
}

func NewPusher(jobName string) *Pusher {
	return &Pusher{
		jobName: jobName,
	}
}

func (p *Pusher) Start() *chan bool {
	ticker := time.NewTicker(3 * time.Second)
	stopChan := make(chan bool)

	log.Println("Starting prometheus pusher")

	go func() {
		for {
			select {
			case <-stopChan:
				log.Println("Stopping prometheus pusher")
				return
			case <-ticker.C:
				p.Push()
			}
		}
	}()

	return &stopChan
}

func (p *Pusher) Push() {
	// Memory
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	memoryAlloc.Set(float64(mem.Alloc))
	memoryTotalAlloc.Set(float64(mem.TotalAlloc))
	memoryHeapAlloc.Set(float64(mem.HeapAlloc))
	memorySys.Set(float64(mem.Sys))

	// Goroutine
	goroutineNum.Set(float64(runtime.NumGoroutine()))

	// Push
	err := push.New(config.GetPushGatewayUrl(), p.jobName).
		Collector(memoryAlloc).
		Collector(memoryTotalAlloc).
		Collector(memoryHeapAlloc).
		Collector(memorySys).
		Collector(goroutineNum).
		Collector(RequestNum).
		Collector(ResponseNum).
		Collector(TaskIdWorkerNum).
		Collector(InputChanNum).
		Collector(OutputChanNum).
		Collector(ResponseTime).
		Add()

	if err != nil {
		fmt.Println("Could not push to push gateway:", err)
	}
}
