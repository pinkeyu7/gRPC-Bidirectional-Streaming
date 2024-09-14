package connectionstatus

import (
	"strings"
	"time"
)

type clientStatus int

const (
	clientStatusStart clientStatus = iota
	clientStatusConnected
	clientStatusDisconnected
	clientStatusRequestChanExisted
)

type ConnectionStatus struct {
	statusChan chan clientStatus
	eventChan  chan struct{}
}

func NewConnectionStatus() *ConnectionStatus {
	c := &ConnectionStatus{
		statusChan: make(chan clientStatus),
		eventChan:  make(chan struct{}),
	}

	c.handleStatusChan()

	return c
}

func (c *ConnectionStatus) Start() {
	c.statusChan <- clientStatusStart
}

func (c *ConnectionStatus) DeferClose() {
	c.statusChan <- clientStatusDisconnected
}

func (c *ConnectionStatus) Connected() {
	c.statusChan <- clientStatusConnected
}

func (c *ConnectionStatus) Error(err error) {
	if strings.Contains(err.Error(), "request channel already exists") {
		c.statusChan <- clientStatusRequestChanExisted
	}
}

func (c *ConnectionStatus) EventChan() chan struct{} {
	return c.eventChan
}

func (c *ConnectionStatus) handleStatusChan() {
	sleepTime := initSleepTime()

	go func() {
		for status := range c.statusChan {
			switch status {
			case clientStatusStart:
				c.eventChan <- struct{}{}
			case clientStatusDisconnected:
				time.Sleep(sleepTime())
				c.eventChan <- struct{}{}
			case clientStatusConnected:
				sleepTime = initSleepTime()
			case clientStatusRequestChanExisted:
				sleepTime = defaultErrorSleepTime
			}
		}
	}()
}
