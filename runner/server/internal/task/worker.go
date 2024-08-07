package task

import (
	taskProto "grpc-bidirectional-streaming/pb/task"
	"grpc-bidirectional-streaming/pkg/prometheus"
	"io"
	"log"

	cmap "github.com/orcaman/concurrent-map/v2"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

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
	s.inputChanMap.Set(workerId, &inputChan)

	outputChanMap := cmap.New[*chan *taskProto.RequestFromServerResponse]()
	s.outputChanMap.Set(workerId, &outputChanMap)

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
			s.inputChanMap.Remove(workerId)
			return nil
		}
		if err != nil {
			s.inputChanMap.Remove(workerId)
			return err
		}

		go func(outputChanMap *cmap.ConcurrentMap[string, *chan *taskProto.RequestFromServerResponse], res *taskProto.RequestFromServerResponse) {
			prometheus.ResponseNum.Inc()

			// Send to output channel
			outputChan, ok := outputChanMap.Get(res.GetRequestId())
			if !ok {
				log.Printf("failed to find output channel: request id: %v", res.GetRequestId())
				return
			}

			// Send to output channel
			//log.Printf("Response: request id: %s, worker id: %s, task id: %s", res.GetRequestId(), workerId, res.GetTaskId())
			*outputChan <- res
			outputChanMap.Remove(res.GetRequestId())
		}(&outputChanMap, res)
	}
}
