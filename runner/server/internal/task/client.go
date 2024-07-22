package task

import (
	"context"
	taskProto "grpc-bidirectional-streaming/pb/task"
	"grpc-bidirectional-streaming/pkg/helper"
	"grpc-bidirectional-streaming/pkg/prometheus"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) RequestFromClient(context context.Context, req *taskProto.RequestFromClientRequest) (*taskProto.RequestFromClientResponse, error) {
	// Check task exist
	workerIdObj, ok := s.taskIdWorkerMap.Load(req.GetTaskId())
	if !ok {
		return nil, status.Error(codes.NotFound, "task not found")
	}
	workerId, ok := workerIdObj.(string)
	if !ok {
		return nil, status.Error(codes.NotFound, "task not found")
	}

	// Monitoring
	start := time.Now()
	prometheus.RequestNum.Inc()

	// Arrange
	requestId := helper.RandString(10)
	log.Printf("Received: request id: %s, worker id: %s, task id: %s", requestId, workerId, req.GetTaskId())

	outputChan := make(chan *taskProto.RequestFromServerResponse)
	defer close(outputChan)

	s.outputChanMap.Store(requestId, &outputChan)

	// Send to input channel
	inputChanObj, ok := s.inputChanMap.Load(workerId)
	if !ok {
		s.outputChanMap.Delete(requestId)
		return nil, status.Error(codes.NotFound, "worker channel not found")
	}
	inputChan, ok := inputChanObj.(*chan *taskProto.RequestFromServerRequest)
	if !ok {
		s.outputChanMap.Delete(requestId)
		return nil, status.Error(codes.NotFound, "worker channel not found")
	}

	reqFromWorker := &taskProto.RequestFromServerRequest{
		RequestId: requestId,
		TaskId:    req.GetTaskId(),
	}
	*inputChan <- reqFromWorker

	// Return
	select {
	case resFromWorker := <-outputChan:
		res := &taskProto.RequestFromClientResponse{
			TaskId:      resFromWorker.GetTaskId(),
			TaskMessage: resFromWorker.GetTaskMessage(),
		}

		duration := time.Since(start)
		prometheus.ResponseTime.WithLabelValues("success").Observe(duration.Seconds())

		return res, nil
	case <-time.After(60 * time.Second):
		duration := time.Since(start)
		prometheus.ResponseTime.WithLabelValues("fail").Observe(duration.Seconds())

		s.outputChanMap.Delete(requestId)
		return nil, status.Errorf(codes.Aborted, "reach timeout")
	}
}
