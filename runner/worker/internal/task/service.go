package task

import (
	"fmt"
	"grpc-bidirectional-streaming/dto/model"
	taskForwardProto "grpc-bidirectional-streaming/pb/task_forward"
	"grpc-bidirectional-streaming/pkg/helper"
)

type Service struct {
	Tasks []model.Task
}

func NewService(workerId string, taskNumber int) *Service {
	// Init tasks
	tasks := make([]model.Task, taskNumber)

	for i := 0; i < taskNumber; i++ {
		taskId := fmt.Sprintf("task_%s_%04d", workerId, i+1)
		tasks[i] = model.Task{
			Id:      taskId,
			Message: helper.Sha1Str(taskId),
		}
	}

	return &Service{
		Tasks: tasks,
	}
}

func (s *Service) HandleRequest(req *taskForwardProto.FooRequest) (*taskForwardProto.FooResponse, error) {
	taskMessage := ""
	for _, t := range s.Tasks {
		if t.Id == req.TaskId {
			taskMessage = t.Message
			break
		}
	}

	return &taskForwardProto.FooResponse{
		RequestId:   req.GetRequestId(),
		TaskId:      req.GetTaskId(),
		TaskMessage: taskMessage,
	}, nil
}
