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

func handleClientStreamRequest[Request any](ctx context.Context, mappingService *MappingService, funcName string, clientId string, req *Request, resChan *chan any, errChan *chan error) {
	// Arrange
	requestId := randString(10)
	err := setFieldValue(req, "RequestId", requestId)
	if err != nil {
		*errChan <- err
		return
	}

	// Monitoring
	start := time.Now()
	prometheus.RequestNum.Inc()

	// Arrange
	log.Printf("Received: request id: %s", requestId)

	// Jaeger
	ctx, span := jaeger.Tracer().Start(ctx, "handle_client_stream_request")
	span.SetAttributes(attribute.String("worker_id", clientId))
	span.SetAttributes(attribute.String("request_id", requestId))
	span.AddEvent("init")
	defer span.End()

	// Get request chan
	requestChan, err := mappingService.GetRequestChan(getPackageNameFromStruct(req), funcName, clientId)
	if err != nil {
		*errChan <- err
		return
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
	for {
		select {
		case resFromWorkerObj := <-responseChan:

			duration := time.Since(start)
			prometheus.ResponseTime.WithLabelValues("success").Observe(duration.Seconds())

			span.AddEvent("response received")

			*resChan <- resFromWorkerObj
		case <-ctx.Done():
			duration := time.Since(start)
			prometheus.ResponseTime.WithLabelValues("fail").Observe(duration.Seconds())

			log.Println("context done - handleClientStreamRequest")
			span.AddEvent("context done")
			return
		}
	}
}

func ForwardClientStreamRequestHandler[Request any, Reply any, ProtoRequest any, ProtoReply any](ctx context.Context, mappingService *MappingService, clientId string, req *Request, resChan *chan *Reply, errChan *chan error) {
	// Jaeger
	ctx, span := jaeger.Tracer().Start(ctx, "forward_client_stream_request_handler")
	span.SetAttributes(attribute.String("worker_id", clientId))
	span.AddEvent("init")
	defer span.End()

	// Arrange
	responseChan := make(chan any)
	defer close(responseChan)

	// Convert dto.request to protobuf.request
	span.AddEvent("convert dto.request to protobuf.request")
	var reqTo ProtoRequest
	err := convert(req, &reqTo)
	if err != nil {
		log.Printf("request marshal: %s", err.Error())
		*errChan <- fmt.Errorf("request marshal failed")
		return
	}

	// Get func name
	funcName := getParentFunctionName(2)

	// Handle request
	span.AddEvent("handle client stream request")
	go handleClientStreamRequest(ctx, mappingService, funcName, clientId, &reqTo, &responseChan, errChan)

	for {
		select {
		case resObj := <-responseChan:
			// Convert protobuf.response
			span.AddEvent("convert protobuf.response")
			res, ok := resObj.(*ProtoReply)
			if !ok {
				log.Printf("convert response failed")
				*errChan <- fmt.Errorf("convert response failed")
				return
			}

			// Get error from response
			span.AddEvent("retrieve error from protobuf.response")
			errFromRes, err := getError(res)
			if err != nil {
				*errChan <- fmt.Errorf("retrieve error: %s", err.Error())
				return
			}
			if errFromRes != nil {
				*errChan <- fmt.Errorf("%s", errFromRes.Message)
				return
			}

			// Convert dto.response
			span.AddEvent("convert protobuf.response to dto.response")
			var resTo Reply
			err = convert(res, &resTo)
			if err != nil {
				log.Printf("reply marshal: %s", err.Error())
				*errChan <- fmt.Errorf("reply marshal failed")
				return
			}

			*resChan <- &resTo
		case <-ctx.Done():
			log.Println("context done - ForwardClientStreamRequestHandler")
			span.AddEvent("context done")
			return
		}
	}
}