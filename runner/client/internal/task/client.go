package task

import (
	"context"
	taskProto "grpc-bidirectional-streaming/pb/task"
	"log"

	"google.golang.org/grpc"
)

type Client struct {
	taskClient taskProto.TaskClient
}

func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{
		taskClient: taskProto.NewTaskClient(conn),
	}
}

func (c *Client) GetInfo(taskId string) error {
	// Arrange
	req := &taskProto.RequestFromClientRequest{
		TaskId: taskId,
	}

	// Act
	res, err := c.taskClient.RequestFromClient(context.Background(), req)
	if err != nil {
		return err
	}

	log.Printf("task id: %s, task message: %s", res.GetTaskId(), res.GetTaskMessage())

	// Return
	return nil
}
