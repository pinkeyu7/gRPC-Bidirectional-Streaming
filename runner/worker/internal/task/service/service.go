package service

import (
	"grpc-bidirectional-streaming/runner/worker/internal/task"
	"log"
)

type Service struct {
}

func NewService() task.Service {
	return &Service{}
}

func (s *Service) GetInfo(taskId string) (string, error) {
	log.Printf("GetInfo, task id: %s", taskId)

	return taskId, nil
}
