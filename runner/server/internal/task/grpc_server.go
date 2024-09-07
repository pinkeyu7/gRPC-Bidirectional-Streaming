package task

import (
	"context"
	"grpc-bidirectional-streaming/config"
	"grpc-bidirectional-streaming/dto"
	taskProto "grpc-bidirectional-streaming/pb/task"
	grpcstreaming "grpc-bidirectional-streaming/pkg/grpc_streaming"
	"grpc-bidirectional-streaming/pkg/helper"
	"grpc-bidirectional-streaming/pkg/jaeger"
	taskForward "grpc-bidirectional-streaming/runner/server/internal/task_forward"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		return nil, errorHandler(err)
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

	errChan := make(chan *grpcstreaming.ErrorInfo)
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
					errChan <- grpcstreaming.NewError(grpcstreaming.ErrorCodeConvertStruct, err.Error())
					return
				}

				err = stream.Send(&res)
				if err != nil {
					log.Printf("send response error: %s", err.Error())
					errChan <- grpcstreaming.NewError(grpcstreaming.ErrorCodeConvertStruct, err.Error())
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
		log.Printf("receive response error: %s", err.Message)
		span.AddEvent("error")
		return errorHandler(err)
	case <-ctx.Done():
		log.Println("context done - grpc")
		span.AddEvent("context done - grpc")
		return nil
	}
}

func errorHandler(err *grpcstreaming.ErrorInfo) error {
	switch err.Code {
	case grpcstreaming.ErrorCodeUnknownError:
		return status.Error(codes.Unknown, err.Message)
	case grpcstreaming.ErrorCodeSendRequest:
		return status.Error(codes.Internal, err.Message)
	case grpcstreaming.ErrorCodeSendResponse:
		return status.Error(codes.Internal, err.Message)
	case grpcstreaming.ErrorCodeSetField:
		return status.Error(codes.Internal, err.Message)
	case grpcstreaming.ErrorCodeRequestChanNotFound:
		return status.Error(codes.Unavailable, err.Message)
	case grpcstreaming.ErrorCodeRequestMarshal:
		return status.Error(codes.Internal, err.Message)
	case grpcstreaming.ErrorCodeConvertStruct:
		return status.Error(codes.Internal, err.Message)
	case grpcstreaming.ErrorCodeRetrieveError:
		return status.Error(codes.Internal, err.Message)
	case grpcstreaming.ErrorCodeReplyMarshal:
		return status.Error(codes.InvalidArgument, err.Message)
	case grpcstreaming.ErrorCodeBadRequest:
		return status.Error(codes.InvalidArgument, err.Message)
	case grpcstreaming.ErrorCodeClientTimeout:
		return status.Error(codes.DeadlineExceeded, err.Message)
	case grpcstreaming.ErrorCodeServerTimeout:
		return status.Error(codes.DeadlineExceeded, err.Message)
	case grpcstreaming.ErrorCodeUnauthorized:
		return status.Error(codes.Unauthenticated, err.Message)
	case grpcstreaming.ErrorCodeForbidden:
		return status.Error(codes.Unauthenticated, err.Message)
	case grpcstreaming.ErrorCodeNotFound:
		return status.Error(codes.NotFound, err.Message)
	}
	return status.Error(codes.Internal, err.Message)
}
