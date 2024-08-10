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

func (c *Client) InitTaskMessage(workerId string, taskNumber int) {
	for i := 0; i < taskNumber; i++ {
		taskId := fmt.Sprintf("task_%s_%04d", workerId, i+1)
		c.taskMessages.Set(taskId, helper.Sha1Str(taskId))
	}
}

func (c *Client) GetInfo(ctx context.Context, workerId string, taskId string) error {
	// Arrange
	req := &taskProto.FooRequest{
		WorkerId: workerId,
		TaskId:   taskId,
	}

	ctx, span := jaeger.Tracer().Start(ctx, "get task")
	span.SetAttributes(attribute.String("task_id", taskId))
	span.AddEvent("send request")
	defer span.End()

	// Act
	res, err := c.taskClient.Foo(ctx, req)
	if err != nil {
		return err
	}

	log.Printf("worker id: %s, task id: %s, task message: %s", res.GetWorkerId(), res.GetTaskId(), res.GetTaskMessage())

	// Validation
	taskMessage, ok := c.taskMessages.Get(taskId)
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
