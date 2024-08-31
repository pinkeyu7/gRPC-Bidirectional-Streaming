package task_forward

import (
	"fmt"
	taskForwardProto "grpc-bidirectional-streaming/pb/task_forward"
	"grpc-bidirectional-streaming/pkg/grpc_streaming"
	"grpc-bidirectional-streaming/pkg/helper"
	"log"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type Service struct {
	taskMessages cmap.ConcurrentMap[string, string]
}

func NewService() *Service {
	return &Service{
		taskMessages: cmap.New[string](),
	}
}

func (s *Service) InitTaskMessage(workerId string, taskNumber int) {
	for i := 0; i < taskNumber; i++ {
		taskId := fmt.Sprintf("task_%s_%04d", workerId, i+1)
		s.taskMessages.Set(taskId, helper.Sha1Str(taskId))
	}
}

func (s *Service) Foo(req *taskForwardProto.FooRequest, resChan *chan *taskForwardProto.FooResponse) {
	// Defer func to prevent sent to close channel
	defer func() {
		if r := recover(); r != nil {
			log.Println("recover from resChan")
		}
	}()

	// Act
	taskMessage, ok := s.taskMessages.Get(req.GetTaskId())
	if !ok {
		*resChan <- grpc_streaming.CreateErrorResponse[taskForwardProto.FooResponse](req.RequestId, 0, "task not found")
	}

	// Return
	*resChan <- &taskForwardProto.FooResponse{
		Error:       nil,
		RequestId:   req.GetRequestId(),
		TaskId:      req.GetTaskId(),
		TaskMessage: taskMessage,
	}
}
