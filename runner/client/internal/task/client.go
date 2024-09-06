package task

import (
	"context"
	"errors"
	"fmt"
	taskProto "grpc-bidirectional-streaming/pb/task"
	"grpc-bidirectional-streaming/pkg/helper"
	"grpc-bidirectional-streaming/pkg/jaeger"
	"log"

	cmap "github.com/orcaman/concurrent-map/v2"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
)

type Client struct {
	taskClient   taskProto.TaskClient
	taskMessages cmap.ConcurrentMap[string, string]
}

func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{
		taskClient:   taskProto.NewTaskClient(conn),
		taskMessages: cmap.New[string](),
	}
}

func (c *Client) InitTaskMessage(workerID string, taskNumber int) {
	for i := 0; i < taskNumber; i++ {
		taskID := fmt.Sprintf("task_%s_%04d", workerID, i+1)
		c.taskMessages.Set(taskID, helper.Sha1Str(taskID))
	}
}

func (c *Client) GetInfo(ctx context.Context, workerID string, taskID string) error {
	// Arrange
	req := &taskProto.UnaryRequest{
		WorkerId: workerID,
		TaskId:   taskID,
	}

	ctx, span := jaeger.Tracer().Start(ctx, "get task")
	span.SetAttributes(attribute.String("task_id", taskID))
	span.AddEvent("send request")
	defer span.End()

	// Act
	res, err := c.taskClient.Unary(ctx, req)
	if err != nil {
		return err
	}

	log.Printf("worker id: %s, task id: %s, task message: %s", res.GetWorkerId(), res.GetTaskId(), res.GetTaskMessage())

	// Validation
	taskMessage, ok := c.taskMessages.Get(taskID)
	if !ok {
		return errors.New("task not found")
	}
	if res.GetTaskMessage() != taskMessage {
		return errors.New("task message error")
	}

	span.AddEvent("receive result")

	// Return
	return nil
}
