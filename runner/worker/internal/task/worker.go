package task

import (
	"context"
	taskProto "grpc-bidirectional-streaming/pb/task"
	"io"
	"log"

	"google.golang.org/grpc"
)

type Client struct {
	taskClient  taskProto.TaskClient
	taskService Service
}

func NewClient(conn *grpc.ClientConn, ts Service) *Client {
	return &Client{
		taskClient:  taskProto.NewTaskClient(conn),
		taskService: ts,
	}
}

func (c *Client) GetInfo() {
	stream, err := c.taskClient.RequestFromServer(context.Background())
	if err != nil {
		log.Fatalf("error: %v", err)
	}

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

		// Act
		taskId := req.GetTaskId()
		taskMessage, err := c.taskService.GetInfo(taskId)

		// Return message
		res := &taskProto.RequestFromServerResponse{
			RequestId:   req.GetRequestId(),
			TaskId:      taskId,
			TaskMessage: taskMessage,
		}
		if err := stream.Send(res); err != nil {
			log.Printf("failed to return: %v", err)
		}
	}
}
