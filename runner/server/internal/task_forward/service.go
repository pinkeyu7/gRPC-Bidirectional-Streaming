package task_forward

import (
	"context"
	"grpc-bidirectional-streaming/config"
	"grpc-bidirectional-streaming/dto"
	taskForwardProto "grpc-bidirectional-streaming/pb/task_forward"
	"grpc-bidirectional-streaming/pkg/grpc_streaming"
	"grpc-bidirectional-streaming/pkg/helper"
	"grpc-bidirectional-streaming/pkg/jaeger"
	"grpc-bidirectional-streaming/pkg/prometheus"
	"log"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	server *Server
}

func NewService(s *Server) *Service {
	return &Service{
		server: s,
	}
}

func (s *Service) Foo(context context.Context, req *dto.FooRequest) (*dto.FooResponse, error) {
	// Monitoring
	start := time.Now()
	prometheus.RequestNum.Inc()

	// Arrange
	requestId := helper.RandString(10)
	log.Printf("Received: request id: %s, worker id: %s, task id: %s", requestId, req.WorkerId, req.TaskId)

	// Jaeger
	_, span := jaeger.Tracer().Start(context, "request_from_client")
	span.SetAttributes(attribute.String("worker_id", req.WorkerId))
	span.SetAttributes(attribute.String("task_id", req.TaskId))
	span.SetAttributes(attribute.String("request_id", requestId))
	span.AddEvent("init")
	defer span.End()

	// Send to input channel
	requestChanIndex := grpc_streaming.GetChanIndex(req.WorkerId, helper.GetCurrentFunctionName())
	requestChan, ok := s.server.RequestChanMap.Get(requestChanIndex)
	if !ok {
		return nil, status.Error(codes.NotFound, "worker channel not found")
	}

	responseChan := make(chan any)
	defer close(responseChan)

	s.server.ResponseChanMap.Set(requestId, &responseChan)

	reqFromWorker := &taskForwardProto.FooRequest{
		RequestId: requestId,
		TaskId:    req.TaskId,
	}
	*requestChan <- reqFromWorker

	span.AddEvent("send to requestChan")

	// Return
	select {
	case resFromWorkerObj := <-responseChan:
		resFromWorker, ok := resFromWorkerObj.(*taskForwardProto.FooResponse)
		if !ok {
			return nil, status.Errorf(codes.Internal, "response convert error")
		}

		res := &dto.FooResponse{
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

		s.server.ResponseChanMap.Remove(requestId)
		return nil, status.Errorf(codes.Aborted, "reach timeout")
	}
}
