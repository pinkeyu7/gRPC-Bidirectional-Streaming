package task

import (
	"context"
	"grpc-bidirectional-streaming/config"
	taskProto "grpc-bidirectional-streaming/pb/task"
	"grpc-bidirectional-streaming/pkg/helper"
	"grpc-bidirectional-streaming/pkg/jaeger"
	"grpc-bidirectional-streaming/pkg/prometheus"
	"log"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) RequestFromClient(context context.Context, req *taskProto.RequestFromClientRequest) (*taskProto.RequestFromClientResponse, error) {
	// Check task exist
	workerId, ok := s.taskIdWorkerMap.Get(req.GetTaskId())
	if !ok {
		return nil, status.Error(codes.NotFound, "task not found")
	}

	// Monitoring
	start := time.Now()
	prometheus.RequestNum.Inc()

	// Arrange
	requestId := helper.RandString(10)
	log.Printf("Received: request id: %s, worker id: %s, task id: %s", requestId, workerId, req.GetTaskId())

	// Jaeger
	_, span := jaeger.Tracer().Start(context, "request_from_client")
	span.SetAttributes(attribute.String("task_id", req.GetTaskId()))
	span.SetAttributes(attribute.String("request_id", requestId))
	span.SetAttributes(attribute.String("worker_id", workerId))
	span.AddEvent("init")
	defer span.End()

	// Send to input channel
	inputChan, ok := s.inputChanMap.Get(workerId)
	if !ok {
		return nil, status.Error(codes.NotFound, "worker channel not found")
	}

	outputChan := make(chan *taskProto.RequestFromServerResponse)
	defer close(outputChan)

	outputChanMap, ok := s.outputChanMap.Get(workerId)
	if !ok {
		return nil, status.Error(codes.NotFound, "worker channel not found")
	}
	outputChanMap.Set(requestId, &outputChan)

	reqFromWorker := &taskProto.RequestFromServerRequest{
		RequestId: requestId,
		TaskId:    req.GetTaskId(),
	}
	*inputChan <- reqFromWorker

	span.AddEvent("send to inputChan")

	// Return
	select {
	case resFromWorker := <-outputChan:
		res := &taskProto.RequestFromClientResponse{
			TaskId:      resFromWorker.GetTaskId(),
			TaskMessage: resFromWorker.GetTaskMessage(),
		}

		duration := time.Since(start)
		prometheus.ResponseTime.WithLabelValues("success").Observe(duration.Seconds())

		span.AddEvent("success")

		return res, nil
	case <-time.After(time.Duration(config.GetServerTimeout()) * time.Second):
		duration := time.Since(start)
		prometheus.ResponseTime.WithLabelValues("fail").Observe(duration.Seconds())

		span.AddEvent("timeout")

		outputChanMap.Remove(requestId)
		return nil, status.Errorf(codes.Aborted, "reach timeout")
	}
}
