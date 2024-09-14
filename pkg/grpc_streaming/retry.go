package grpcstreaming

import (
	"context"
	"log"
	"strings"
	"time"
)

type clientStatus int

const (
	defaultSleepTime = 10

	clientConnected clientStatus = iota
	clientRequestChanExisted
)

type connectionAgent struct {
	reconnectChan chan struct{}
	statusChan    chan clientStatus
}

func newConnectionAgent(ctx context.Context, function func(ctx context.Context)) *connectionAgent {
	c := &connectionAgent{
		reconnectChan: make(chan struct{}),
		statusChan:    make(chan clientStatus),
	}

	go c.handleChannel(ctx, function)

	return c
}

func (c *connectionAgent) Start() {
	c.reconnectChan <- struct{}{}
}

func (c *connectionAgent) Disconnected() {
	c.reconnectChan <- struct{}{}
}

func (c *connectionAgent) Connected() {
	c.statusChan <- clientConnected
}

func (c *connectionAgent) Error(err error) {
	if strings.Contains(err.Error(), "request channel already exists") {
		c.statusChan <- clientRequestChanExisted
	}
}

func (c *connectionAgent) handleChannel(ctx context.Context, function func(ctx context.Context)) {
	sleepTime := initSleepTime()

	go func() {
		for {
			select {
			case <-c.reconnectChan:
				time.Sleep(sleepTime())
				go function(ctx)
			case status := <-c.statusChan:
				switch status {
				case clientConnected:
					sleepTime = initSleepTime()
				case clientRequestChanExisted:
					sleepTime = defaultErrorSleepTime
				default:
					return
				}
			}
		}
	}()
}

var fibonacciDelays = []int{1, 1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144, 233, 300}

func initSleepTime() func() time.Duration {
	i := 0
	return func() time.Duration {
		delayDuration := 0

		if i < len(fibonacciDelays) {
			delayDuration = fibonacciDelays[i]
		} else {
			delayDuration = fibonacciDelays[len(fibonacciDelays)-1]
		}

		i++

		timeDuration := time.Duration(delayDuration) * 100 * time.Millisecond
		log.Printf("Trying %d times; waiting %v ms before retrying...\n", i, timeDuration.Milliseconds())

		return timeDuration
	}
}

func defaultErrorSleepTime() time.Duration {
	return defaultSleepTime * time.Second
}
