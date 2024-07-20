package task

import (
	"context"
	taskProto "grpc-bidirectional-streaming/pb/task"
	"grpc-bidirectional-streaming/pkg/helper"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) RequestFromClient(context context.Context, req *taskProto.RequestFromClientRequest) (*taskProto.RequestFromClientResponse, error) {
	// Arrange
	requestId := helper.RandString(10)
	log.Printf("Received: request id: %s, task id: %s", requestId, req.GetTaskId())

	outputChan := make(chan *taskProto.RequestFromServerResponse)
	s.outputChanMap.Store(requestId, &outputChan)

	// Setup timeout
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(2 * time.Second)
		timeout <- true
		close(timeout)
	}()

	// Send to input channel
	inputChanObj, ok := s.inputChanMap.Load("worker")
	if !ok {
		log.Printf("failed to find input channel: worker: %v", "worker")
	} else {
		inputChan, ok := inputChanObj.(*chan *taskProto.RequestFromServerRequest)
		if !ok {
			log.Printf("failed to find input channel: worker: %v", "worker")
		} else {
			reqFromWorker := &taskProto.RequestFromServerRequest{
				RequestId: requestId,
				TaskId:    req.GetTaskId(),
			}

			*inputChan <- reqFromWorker
		}
	}

	// Return
	select {
	case resFromWorker := <-outputChan:
		res := &taskProto.RequestFromClientResponse{
			TaskId:      resFromWorker.GetTaskId(),
			TaskMessage: resFromWorker.GetTaskMessage(),
		}

		return res, nil
	case <-timeout:
		return nil, status.Errorf(codes.Aborted, "timeout")
	}
}
