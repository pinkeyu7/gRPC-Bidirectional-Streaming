package grpc_streaming

import (
	"context"
	"fmt"
	"grpc-bidirectional-streaming/pkg/jaeger"
	"grpc-bidirectional-streaming/pkg/prometheus"
	"log"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

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

		return nil, fmt.Errorf("server - request timeout")
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
