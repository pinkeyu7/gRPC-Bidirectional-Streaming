package task

import (
	taskProto "grpc-bidirectional-streaming/pb/task"
	"grpc-bidirectional-streaming/pkg/prometheus"
	"io"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"context"
)

func (s *Server) RegisterFromWorker(context context.Context, req *taskProto.RegisterFromWorkerRequest) (*taskProto.RegisterFromWorkerResponse, error) {
	for _, taskId := range req.GetTaskIds() {
		s.taskIdWorkerMap.Store(taskId, req.GetWorkerId())
	}

	log.Printf("RegisterFromWorker worker id: %v", req.GetWorkerId())

	return &taskProto.RegisterFromWorkerResponse{}, nil
}

func (s *Server) UnregisterFromWorker(workerId string) {
	s.taskIdWorkerMap.Range(func(key, workerIdObj interface{}) bool {
		if workerIdObj.(string) == workerId {
			s.taskIdWorkerMap.Delete(key)
		}
		return true
	})

	log.Printf("UnregisterFromWorker worker id: %v", workerId)
}

func (s *Server) RequestFromServer(stream taskProto.Task_RequestFromServerServer) error {
	// Read metadata from client
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return status.Errorf(codes.DataLoss, "failed to get metadata")
	}
	workerId := ""
	if rs, ok := md["worker_id"]; ok {
		for _, wid := range rs {
			workerId = wid
		}
	}

	// Arrange
	inputChan := make(chan *taskProto.RequestFromServerRequest)
	defer close(inputChan)
	s.inputChanMap.Store(workerId, &inputChan)

	// Request from client, send to worker
	go func() {
		for req := range inputChan {
			if err := stream.Send(req); err != nil {
				log.Printf("failed to send request: %v", err)
			}
		}
	}()

	// Receive from worker, send to client
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			s.UnregisterFromWorker(workerId)
			s.inputChanMap.Delete(workerId)
			return nil
		}
		if err != nil {
			s.UnregisterFromWorker(workerId)
			s.inputChanMap.Delete(workerId)
			return err
		}

		// Send to output channel
		outputChanObj, ok := s.outputChanMap.Load(res.GetRequestId())
		if !ok {
			log.Printf("failed to find output channel: request id: %v", res.GetRequestId())
			continue
		}
		outputChan, ok := outputChanObj.(*chan *taskProto.RequestFromServerResponse)
		if !ok {
			log.Printf("failed to find output channel: request id: %v", res.GetRequestId())
			s.outputChanMap.Delete(res.GetRequestId())
			continue
		}

		// Send to output channel
		log.Printf("Response: request id: %s, worker id: %s, task id: %s", res.GetRequestId(), workerId, res.GetTaskId())
		*outputChan <- res
		s.outputChanMap.Delete(res.GetRequestId())

		prometheus.RequestNum.Add(float64(-1))
	}
}
