package task_forward

import (
	"errors"
	"fmt"
	taskForwardProto "grpc-bidirectional-streaming/pb/task_forward"
	"grpc-bidirectional-streaming/pkg/helper"

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

func (s *Service) HandleRequest(req *taskForwardProto.FooRequest) (*taskForwardProto.FooResponse, error) {
	taskMessage, ok := s.taskMessages.Get(req.GetTaskId())
	if !ok {
		return nil, errors.New("task not found")
	}

	return &taskForwardProto.FooResponse{
		RequestId:   req.GetRequestId(),
		TaskId:      req.GetTaskId(),
		TaskMessage: taskMessage,
	}, nil
}
