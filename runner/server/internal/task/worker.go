package task

import (
	taskProto "grpc-bidirectional-streaming/pb/task"
	"io"
	"log"
)

func (s *Server) RequestFromServer(stream taskProto.Task_RequestFromServerServer) error {
	// Arrange
	inputChan := make(chan *taskProto.RequestFromServerRequest)
	s.inputChanMap.Store("worker", &inputChan)

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
			s.inputChanMap.Delete("worker")
			close(inputChan)
			return nil
		}
		if err != nil {
			s.inputChanMap.Delete("worker")
			close(inputChan)
			return err
		}

		// Send to output channel
		outputChanObj, ok := s.outputChanMap.Load(res.RequestId)
		if !ok {
			log.Printf("failed to find output channel: request id: %v", res.GetRequestId())
		} else {
			outputChan, ok := outputChanObj.(*chan *taskProto.RequestFromServerResponse)
			if !ok {
				log.Printf("failed to find output channel: request id: %v", res.GetRequestId())
			} else {
				*outputChan <- res
				s.outputChanMap.Delete(res.GetRequestId())
				close(*outputChan)
			}
		}
	}
}
