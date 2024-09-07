package grpcstreaming

import (
	"context"
	"grpc-bidirectional-streaming/pkg/jaeger"
	"grpc-bidirectional-streaming/pkg/prometheus"
	"log"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

func handleUnaryRequest[Request any](ctx context.Context, mappingService *MappingService, clientID string, req *Request) (any, *ErrorInfo) {
	// Arrange
	requestID := getKsuID()
	err := setFieldValue(req, "RequestId", requestID)
	if err != nil {
		return nil, NewError(ErrorCodeSetField, err.Error())
	}

	// Monitoring
	start := time.Now()
	prometheus.RequestNum.Inc()

	// Arrange
	log.Printf("Received: request id: %s", requestID)

	// Jaeger
	ctx, span := jaeger.Tracer().Start(ctx, "handle_unary_request")
	span.SetAttributes(attribute.String("worker_id", clientID))
	span.SetAttributes(attribute.String("request_id", requestID))
	span.AddEvent("init")
	defer span.End()

	// Get request chan
	requestChan, err := mappingService.GetRequestChan(getPackageNameFromStruct(req), getParentFunctionName(3), clientID)
	if err != nil {
		return nil, NewError(ErrorCodeRequestChanNotFound, err.Error())
	}

	// Set response chan
	responseChan := make(chan any)
	defer func() {
		mappingService.RemoveResponseChan(requestID)
		close(responseChan)
	}()
	mappingService.SetResponseChan(requestID, responseChan)

	// Send request
	requestChan <- req

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

		return nil, NewError(ErrorCodeServerTimeout, "server - request timeout")
	}
}

func ForwardUnaryRequestHandler[Request any, Reply any, ProtoRequest any, ProtoReply any](
	ctx context.Context, mappingService *MappingService, clientID string, req *Request) (*Reply, *ErrorInfo) {

	// Jaeger
	ctx, span := jaeger.Tracer().Start(ctx, "forward_unary_request_handler")
	span.SetAttributes(attribute.String("worker_id", clientID))
	span.AddEvent("init")
	defer span.End()

	// Convert dto.request to protobuf.request
	span.AddEvent("convert dto.request to protobuf.request")
	var reqTo ProtoRequest
	err := convert(req, &reqTo)
	if err != nil {
		log.Printf("request marshal: %s", err.Error())
		return nil, NewError(ErrorCodeRequestMarshal, err.Error())
	}

	// Handle request
	span.AddEvent("handle unary request")
	resObj, errInfo := handleUnaryRequest(ctx, mappingService, clientID, &reqTo)
	if errInfo != nil {
		log.Printf("handle request: %s", errInfo.Message)
		return nil, errInfo
	}

	// Convert protobuf.response
	span.AddEvent("convert protobuf.response")
	res, ok := resObj.(*ProtoReply)
	if !ok {
		log.Printf("convert response failed")
		return nil, NewError(ErrorCodeConvertStruct, "convert response failed")
	}

	// Get error from response
	span.AddEvent("retrieve error from protobuf.response")
	errFromRes, err := getError(res)
	if err != nil {
		return nil, NewError(ErrorCodeRetrieveError, err.Error())
	}
	if errFromRes != nil {
		return nil, errFromRes
	}

	// Convert dto.response
	span.AddEvent("convert protobuf.response to dto.response")
	var resTo Reply
	err = convert(res, &resTo)
	if err != nil {
		log.Printf("reply marshal: %s", err.Error())
		return nil, NewError(ErrorCodeReplyMarshal, err.Error())
	}

	return &resTo, nil
}
