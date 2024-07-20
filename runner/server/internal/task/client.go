package task

import (
	"context"
	taskProto "grpc-bidirectional-streaming/pb/task"
	"grpc-bidirectional-streaming/pkg/helper"
	"log"
)

func (s *Server) RequestFromClient(context context.Context, req *taskProto.RequestFromClientRequest) (*taskProto.RequestFromClientResponse, error) {
	// Arrange
	requestId := helper.RandString(10)
	log.Printf("Received: request id: %s, task id: %s", requestId, req.GetTaskId())

	// Act
	reqFromWorker := &taskProto.RequestFromServerRequest{
		RequestId: requestId,
		TaskId:    req.GetTaskId(),
	}
	s.inputChan <- reqFromWorker

	// Return
	select {
	case resFromWorker := <-s.outputChan:
		res := &taskProto.RequestFromClientResponse{
			TaskId:      resFromWorker.GetTaskId(),
			TaskMessage: resFromWorker.GetTaskMessage(),
		}

		return res, nil
	}
}
