package task

import (
	"context"
	"grpc-bidirectional-streaming/config"
	"grpc-bidirectional-streaming/dto"
	taskProto "grpc-bidirectional-streaming/pb/task"
	"grpc-bidirectional-streaming/pkg/helper"
	"grpc-bidirectional-streaming/pkg/jaeger"
	"grpc-bidirectional-streaming/runner/server/internal/task_forward"
	"log"
	"time"
)

type Server struct {
	taskProto.UnimplementedTaskServer
	taskForwardService *task_forward.Service
}

func NewServer(tfs *task_forward.Service) *Server {
	return &Server{taskForwardService: tfs}
}

func (s *Server) Foo(ctx context.Context, req *taskProto.FooRequest) (*taskProto.FooResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, config.GetServerTimeout()*time.Second)
	defer cancel()

	// Jaeger
	ctx, span := jaeger.Tracer().Start(ctx, "receive request")
	span.AddEvent("init")
	defer span.End()

	request := &dto.FooRequest{
		WorkerId: req.GetWorkerId(),
		TaskId:   req.TaskId,
	}

	response, err := s.taskForwardService.Foo(ctx, request)
	if err != nil {
		return nil, err
	}

	span.AddEvent("done")

	return &taskProto.FooResponse{
		WorkerId:    response.WorkerId,
		TaskId:      response.TaskId,
		TaskMessage: response.TaskMessage,
	}, nil
}

func (s *Server) UpnpSearchExample(req *taskProto.UpnpSearchRequest, stream taskProto.Task_UpnpSearchExampleServer) error {
	// Arrange context
	ctx, cancel := context.WithTimeout(stream.Context(), config.GetServerTimeout()*time.Second)
	defer cancel()

	// Jaeger
	ctx, span := jaeger.Tracer().Start(ctx, "receive request")
	span.AddEvent("init")
	defer span.End()

	// Init req
	reqTo := &dto.UpnpSearchRequest{
		WorkerId: req.GetWorkerId(),
	}

	// Init response chan
	responseChan := make(chan *dto.UpnpSearchResponse)
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

				var res taskProto.UpnpSearchResponse
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
	go s.taskForwardService.UpnpSearch(ctx, reqTo, &responseChan, &errChan)

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
