package service

import (
	"grpc-bidirectional-streaming/pkg/helper"
	"grpc-bidirectional-streaming/runner/worker/internal/task"
)

type Service struct {
}

func NewService() task.Service {
	return &Service{}
}

func (s *Service) GetInfo(taskId string) (string, error) {
	return helper.Sha1Str(taskId), nil
}
