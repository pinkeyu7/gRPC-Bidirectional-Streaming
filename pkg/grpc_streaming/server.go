package grpc_streaming

import (
	"context"
	"fmt"
	"grpc-bidirectional-streaming/pkg/jaeger"
	"grpc-bidirectional-streaming/pkg/prometheus"
	"io"
	"log"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"google.golang.org/grpc/metadata"
)

type unaryServerObject[Request any, Response any] interface {
	Send(*Request) error
	Recv() (*Response, error)
	Context() context.Context
}

type unaryServer[Request any, Response any, Stream unaryServerObject[Request, Response]] struct {
	clientId       string
	funcName       string
	mappingService *MappingService
	stream         unaryServerObject[Request, Response]
}

func NewUnaryServer[Request any, Response any, Stream unaryServerObject[Request, Response]](ms *MappingService, stream Stream) error {
	s := &unaryServer[Request, Response, Stream]{
		funcName:       getParentFunctionName(2),
		mappingService: ms,
		stream:         stream,
	}

	err := s.setClientId(stream.Context())
	if err != nil {
		return err
	}

	err = s.handleStream()
	if err != nil {
		return err
	}

	return nil
}

func (s *unaryServer[Request, Response, Stream]) setClientId(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return fmt.Errorf("could not extract grpc metadata")
	}
	if rs, ok := md["client_id"]; ok {
		for _, cid := range rs {
			s.clientId = cid
			return nil
		}
	}

	return fmt.Errorf("could not extract grpc client id")
}

func (s *unaryServer[Request, Response, Stream]) handleStream() error {
	// Get caller package name
	var reqStruct Request
	packageName := getPackageNameFromStruct(reqStruct)

	// Arrange
	requestChan := make(chan any)
	defer close(requestChan)

	s.mappingService.SetRequestChan(packageName, s.funcName, s.clientId, &requestChan)

	// Request from client, send to worker
	go func() {
		for reqObj := range requestChan {
			req, ok := reqObj.(*Request)
			if !ok {
				log.Printf("failed to convert request")
				continue
			}

			if err := s.stream.Send(req); err != nil {
				log.Printf("failed to send request: %s", err.Error())
			}
		}
	}()

	// Receive from worker, send to client
	for {
		res, err := s.stream.Recv()
		if err == io.EOF {
			s.mappingService.RemoveRequestChan(packageName, s.funcName, s.clientId)
			return nil
		}
		if err != nil {
			s.mappingService.RemoveRequestChan(packageName, s.funcName, s.clientId)
			return err
		}

		go func(res *Response) {
			// Get requestId
			requestId, err := getFieldValue(res, "RequestId")
			if err != nil {
				log.Printf("failed to print request id: %s", err.Error())
				return
			}

			// Send to output channel
			responseChan, err := s.mappingService.GetResponseChan(requestId)
			if err != nil {
				log.Printf("request id: %s, error: %s", requestId, err.Error())
				return
			}

			// Send to output channel
			*responseChan <- res
		}(res)
	}
}

func handleUnaryRequest[Request any](ctx context.Context, mappingService *MappingService, clientId string, req *Request) (any, error) {
	// Arrange
	requestId := randString(10)
	err := setFieldValue(req, "RequestId", requestId)
	if err != nil {
		return nil, err
	}

	// Monitoring
	start := time.Now()
	prometheus.RequestNum.Inc()

	// Arrange
	log.Printf("Received: request id: %s", requestId)

	// Jaeger
	ctx, span := jaeger.Tracer().Start(ctx, "handle_unary_request")
	span.SetAttributes(attribute.String("worker_id", clientId))
	span.SetAttributes(attribute.String("request_id", requestId))
	span.AddEvent("init")
	defer span.End()

	// Get request chan
	requestChan, err := mappingService.GetRequestChan(getPackageNameFromStruct(req), getParentFunctionName(3), clientId)
	if err != nil {
		return nil, err
	}

	// Set response chan
	responseChan := make(chan any)
	defer func() {
		mappingService.RemoveResponseChan(requestId)
		close(responseChan)
	}()
	mappingService.SetResponseChan(requestId, &responseChan)

	// Send request
	*requestChan <- req

	span.AddEvent("send to requestChan")

	// Return
	select {
	case resFromWorkerObj := <-responseChan:

		duration := time.Since(start)
		prometheus.ResponseTime.WithLabelValues("success").Observe(duration.Seconds())

		span.AddEvent("success")

		return resFromWorkerObj, nil
	case <-ctx.Done():
		duration := time.Since(start)
		prometheus.ResponseTime.WithLabelValues("fail").Observe(duration.Seconds())

		span.AddEvent("timeout")

		return nil, fmt.Errorf("request timeout")
	}
}

func ForwardUnaryRequestHandler[Request any, Reply any, ProtoRequest any, ProtoReply any](ctx context.Context, mappingService *MappingService, clientId string, req *Request) (*Reply, error) {
	// Jaeger
	ctx, span := jaeger.Tracer().Start(ctx, "forward_unary_request_handler")
	span.SetAttributes(attribute.String("worker_id", clientId))
	span.AddEvent("init")
	defer span.End()

	// Convert dto.request to protobuf.request
	span.AddEvent("convert dto.request to protobuf.request")
	var reqTo ProtoRequest
	err := convert(req, &reqTo)
	if err != nil {
		log.Printf("request marshal: %s", err.Error())
		return nil, fmt.Errorf("request marshal failed")
	}

	// Handle request
	span.AddEvent("handle unary request")
	resObj, err := handleUnaryRequest(ctx, mappingService, clientId, &reqTo)
	if err != nil {
		log.Printf("handle request: %s", err.Error())
		return nil, err
	}

	// Convert protobuf.response
	span.AddEvent("convert protobuf.response")
	res, ok := resObj.(*ProtoReply)
	if !ok {
		log.Printf("convert response failed")
		return nil, fmt.Errorf("convert response failed")
	}

	// Get error from response
	span.AddEvent("retrieve error from protobuf.response")
	errFromRes, err := getError(res)
	if err != nil {
		return nil, fmt.Errorf("retrieve error: %s", err.Error())
	}
	if errFromRes != nil {
		return nil, fmt.Errorf("%s", errFromRes.Message)
	}

	// Convert dto.response
	span.AddEvent("convert protobuf.response to dto.response")
	var resTo Reply
	err = convert(res, &resTo)
	if err != nil {
		log.Printf("reply marshal: %s", err.Error())
		return nil, fmt.Errorf("reply marshal failed")
	}

	return &resTo, nil
}
