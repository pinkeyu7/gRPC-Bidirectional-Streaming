package service

import (
	"fmt"
	"grpc-bidirectional-streaming/dto/model"
	"grpc-bidirectional-streaming/pkg/helper"
	"grpc-bidirectional-streaming/runner/worker/internal/task"
)

type Service struct {
	Tasks []model.Task
}

func NewService(taskNumber int) task.Service {
	// Init tasks
	tasks := make([]model.Task, taskNumber)

	for i := 0; i < taskNumber; i++ {
		taskId := fmt.Sprintf("task_%04d", i+1)
		tasks[i] = model.Task{
			Id:      taskId,
			Message: helper.Sha1Str(taskId),
		}
	}

	return &Service{
		Tasks: tasks,
	}
}

func (s *Service) GetInfo(taskId string) (string, error) {
	for _, t := range s.Tasks {
		if t.Id == taskId {
			return t.Message, nil
		}
	}

	return "", nil
}
