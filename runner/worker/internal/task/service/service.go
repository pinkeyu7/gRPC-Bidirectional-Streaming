package service

import (
	"fmt"
	"grpc-bidirectional-streaming/config"
	"grpc-bidirectional-streaming/dto/model"
	"grpc-bidirectional-streaming/pkg/helper"
	"grpc-bidirectional-streaming/runner/worker/internal/task"
	"math/rand/v2"
	"time"
)

type Service struct {
	Tasks []model.Task
}

func NewService(workerId string, taskNumber int) task.Service {
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

func (s *Service) GetIds() []string {
	taskIds := make([]string, len(s.Tasks))
	for i, t := range s.Tasks {
		taskIds[i] = t.Id
	}

	return taskIds
}

func (s *Service) GetInfo(taskId string) (string, error) {
	// Make worker idle
	time.Sleep(time.Duration(rand.IntN(config.GetWorkerIdleTime())) * time.Second)

	// Find task
	for _, t := range s.Tasks {
		if t.Id == taskId {
			return t.Message, nil
		}
	}

	return "", nil
}
