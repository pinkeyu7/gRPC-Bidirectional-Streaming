package task

import (
	taskProto "grpc-bidirectional-streaming/pb/task"
	"io"
	"log"
)

func (s *Server) RequestFromServer(stream taskProto.Task_RequestFromServerServer) error {
	go func() {
		for req := range s.inputChan {
			if err := stream.Send(req); err != nil {
				log.Printf("failed to send request: %v", err)
			}
		}
	}()

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		s.outputChan <- res
	}
}
