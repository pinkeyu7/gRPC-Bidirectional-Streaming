package task

import (
	"context"
	"grpc-bidirectional-streaming/config"
	"grpc-bidirectional-streaming/dto"
	taskProto "grpc-bidirectional-streaming/pb/task"
	"grpc-bidirectional-streaming/pkg/helper"
	"grpc-bidirectional-streaming/pkg/jaeger"
	taskForward "grpc-bidirectional-streaming/runner/server/internal/task_forward"
	"log"
	"time"
)

type Server struct {
	taskProto.UnimplementedTaskServer
	taskForwardService *taskForward.Service
}

func NewServer(tfs *taskForward.Service) *Server {
	return &Server{taskForwardService: tfs}
}

func (s *Server) Unary(ctx context.Context, req *taskProto.UnaryRequest) (*taskProto.UnaryResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, config.GetServerTimeout()*time.Second)
	defer cancel()

	// Jaeger
	ctx, span := jaeger.Tracer().Start(ctx, "receive request")
	span.AddEvent("init")
	defer span.End()

	request := &dto.UnaryRequest{
		WorkerID: req.GetWorkerId(),
		TaskID:   req.TaskId,
	}

	response, err := s.taskForwardService.Unary(ctx, request)
	if err != nil {
		return nil, err
	}

	span.AddEvent("done")

	return &taskProto.UnaryResponse{
		WorkerId:    response.WorkerID,
		TaskId:      response.TaskID,
		TaskMessage: response.TaskMessage,
	}, nil
}

func (s *Server) ClientStream(req *taskProto.ClientStreamRequest, stream taskProto.Task_ClientStreamServer) error {
	// Arrange context
	ctx, cancel := context.WithTimeout(stream.Context(), config.GetServerTimeout()*time.Second)
	defer cancel()

	// Jaeger
	ctx, span := jaeger.Tracer().Start(ctx, "receive request")
	span.AddEvent("init")
	defer span.End()

	// Init req
	reqTo := &dto.ClientStreamRequest{
		WorkerID: req.GetWorkerId(),
	}

	// Init response chan
	responseChan := make(chan *dto.ClientStreamResponse)
	defer close(responseChan)

	errChan := make(chan error)
	defer close(errChan)

	// Handle response
	go func() {
		for {
			select {
			case response, ok := <-responseChan:
				if !ok {
					return
				}

				var res taskProto.ClientStreamResponse
				err := helper.Convert(response, &res)
				if err != nil {
					log.Printf("convert response error: %s", err.Error())
					errChan <- err
					return
				}

				err = stream.Send(&res)
				if err != nil {
					log.Printf("send response error: %s", err.Error())
					errChan <- err
					return
				}
			case <-ctx.Done():
				log.Println("context done - grpc - act func")
				return
			}
		}
	}()

	// Act
	go s.taskForwardService.ClientStream(ctx, reqTo, responseChan, errChan)

	// Wait for response
	select {
	case err := <-errChan:
		log.Printf("receive response error: %s", err.Error())
		span.AddEvent("error")
		return err
	case <-ctx.Done():
		log.Println("context done - grpc")
		span.AddEvent("context done - grpc")
		return nil
	}
}
