package task

import (
	"context"
	taskProto "grpc-bidirectional-streaming/pb/task"
	"grpc-bidirectional-streaming/pkg/jaeger"
	"log"

	"go.opentelemetry.io/otel/attribute"

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

func (c *Client) GetInfo(ctx context.Context, taskId string) error {
	// Arrange
	req := &taskProto.RequestFromClientRequest{
		TaskId: taskId,
	}

	ctx, span := jaeger.Tracer().Start(ctx, "get task")
	span.SetAttributes(attribute.String("task_id", taskId))
	span.AddEvent("send request")
	defer span.End()

	// Act
	res, err := c.taskClient.RequestFromClient(ctx, req)
	if err != nil {
		return err
	}

	log.Printf("task id: %s, task message: %s", res.GetTaskId(), res.GetTaskMessage())

	span.AddEvent("receive result")

	// Return
	return nil
}
