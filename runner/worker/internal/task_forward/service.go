package task_forward

import (
	"context"
	"fmt"
	taskForwardProto "grpc-bidirectional-streaming/pb/task_forward"
	"grpc-bidirectional-streaming/pkg/grpc_streaming"
	"grpc-bidirectional-streaming/pkg/helper"
	"log"
	"time"

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

func (s *Service) Unary(ctx context.Context, req *taskForwardProto.UnaryRequest, resChan *chan *taskForwardProto.UnaryResponse) {
	// Defer func to prevent sent to close channel
	defer func() {
		if r := recover(); r != nil {
			log.Println("recover from resChan")
		}
	}()

	// Act
	taskMessage, ok := s.taskMessages.Get(req.GetTaskId())
	if !ok {
		*resChan <- grpc_streaming.NewErrorResponse[taskForwardProto.UnaryResponse](req.RequestId, grpc_streaming.ErrorCodeNotFound, "task not found")
	}

	// Return
	*resChan <- &taskForwardProto.UnaryResponse{
		Error:       nil,
		RequestId:   req.GetRequestId(),
		TaskId:      req.GetTaskId(),
		TaskMessage: taskMessage,
	}
}

func (s *Service) ClientStream(ctx context.Context, req *taskForwardProto.ClientStreamRequest, resChan *chan *taskForwardProto.ClientStreamResponse) {
	// Defer func to prevent sent to close channel
	defer func() {
		if r := recover(); r != nil {
			log.Println("recover from resChan")
		}
	}()

	// Arrange
	resultChan := make(chan *taskForwardProto.ClientStreamResponse)
	defer close(resultChan)

	// Mock upnp search result
	go func() {
		for i := 0; i < 30; i++ {
			// Arrange
			res := &taskForwardProto.ClientStreamResponse{
				Error:     nil,
				RequestId: req.GetRequestId(),
				Model:     fmt.Sprintf("upnp_search_%d", i),
				Ip:        fmt.Sprintf("192.168.1.%d", i),
			}

			// Send response
			select {
			case <-ctx.Done():
				log.Println("context done - ClientStream - mock")
				return
			default:
				resultChan <- res
			}
			time.Sleep(1 * time.Second)
		}
	}()

	// Send response to resChan
	for result := range resultChan {
		select {
		case <-ctx.Done():
			log.Println("context done - ClientStream")
			return
		default:
			*resChan <- result
		}
	}
}
