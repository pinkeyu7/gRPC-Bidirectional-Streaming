package task

import (
	"context"
	taskProto "grpc-bidirectional-streaming/pb/task"
	"grpc-bidirectional-streaming/pkg/prometheus"
	"io"
	"log"

	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc"
)

type Client struct {
	workerId    string
	taskClient  taskProto.TaskClient
	taskService Service
}

func NewClient(workerId string, conn *grpc.ClientConn, ts Service) *Client {
	return &Client{
		workerId:    workerId,
		taskClient:  taskProto.NewTaskClient(conn),
		taskService: ts,
	}
}

func (c *Client) RegisterIds() error {
	// Arrange
	req := &taskProto.RegisterFromWorkerRequest{
		WorkerId: c.workerId,
		TaskIds:  c.taskService.GetIds(),
	}

	// Act
	_, err := c.taskClient.RegisterFromWorker(context.Background(), req)
	if err != nil {
		return err
	}

	log.Printf("RegisterIds success")

	// Return
	return nil
}

func (c *Client) GetInfo() {
	// Arrange
	outputChan := make(chan *taskProto.RequestFromServerResponse)
	defer close(outputChan)

	// Create metadata and context
	md := metadata.Pairs("worker_id", c.workerId)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// Make RPC using the context with the metadata
	stream, err := c.taskClient.RequestFromServer(ctx)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// Send to server
	go func() {
		for req := range outputChan {
			if err := stream.Send(req); err != nil {
				log.Printf("failed to send request: %v", err)
			}
			prometheus.ResponseNum.Inc()
		}
	}()

	// Handle message
	for {
		// Receive message
		req, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatalf("failed to receive request: %v", err)
		}

		prometheus.RequestNum.Inc()

		go func(req *taskProto.RequestFromServerRequest) {
			prometheus.WorkerRequestNum.Inc()

			// Act
			taskMessage, _ := c.taskService.GetInfo(req.GetTaskId())
			log.Printf("request id: %s, task id: %s, task message: %s", req.GetRequestId(), req.GetTaskId(), taskMessage)

			// Make worker idle
			//time.Sleep(time.Duration(rand.IntN(config.GetWorkerIdleTime())) * time.Second)

			// Return message
			res := &taskProto.RequestFromServerResponse{
				RequestId:   req.GetRequestId(),
				TaskId:      req.GetTaskId(),
				TaskMessage: taskMessage,
			}
			outputChan <- res
		}(req)
	}
}
