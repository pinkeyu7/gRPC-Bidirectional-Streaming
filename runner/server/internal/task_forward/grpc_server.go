package task_forward

import (
	taskForwardProto "grpc-bidirectional-streaming/pb/task_forward"
	"grpc-bidirectional-streaming/pkg/prometheus"
	"io"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type Server struct {
	taskForwardProto.UnimplementedTaskForwardServer
	RequestChanMap  cmap.ConcurrentMap[string, *chan *taskForwardProto.FooRequest]
	ResponseChanMap cmap.ConcurrentMap[string, *chan *taskForwardProto.FooResponse]
}

func NewServer() *Server {
	return &Server{
		RequestChanMap:  cmap.New[*chan *taskForwardProto.FooRequest](),
		ResponseChanMap: cmap.New[*chan *taskForwardProto.FooResponse](),
	}
}

func (s *Server) Monitor() {
	prometheus.RequestChanNum.Set(float64(s.RequestChanMap.Count()))
	prometheus.ResponseChanNum.Set(float64(s.ResponseChanMap.Count()))
}

func (s *Server) Foo(stream taskForwardProto.TaskForward_FooServer) error {
	// Read metadata from client
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return status.Errorf(codes.DataLoss, "failed to get metadata")
	}
	workerId := ""
	if rs, ok := md["client_id"]; ok {
		for _, wid := range rs {
			workerId = wid
		}
	}

	// Arrange
	RequestChan := make(chan *taskForwardProto.FooRequest)
	defer close(RequestChan)
	s.RequestChanMap.Set(workerId, &RequestChan)

	// Request from client, send to worker
	go func() {
		for req := range RequestChan {
			if err := stream.Send(req); err != nil {
				log.Printf("failed to send request: %v", err)
			}
		}
	}()

	// Receive from worker, send to client
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			s.RequestChanMap.Remove(workerId)
			return nil
		}
		if err != nil {
			s.RequestChanMap.Remove(workerId)
			return err
		}

		go func(ResponseChanMap *cmap.ConcurrentMap[string, *chan *taskForwardProto.FooResponse], res *taskForwardProto.FooResponse) {
			prometheus.ResponseNum.Inc()

			// Send to output channel
			ResponseChan, ok := ResponseChanMap.Get(res.GetRequestId())
			if !ok {
				log.Printf("failed to find output channel: request id: %v", res.GetRequestId())
				return
			}

			// Send to output channel
			//log.Printf("Response: request id: %s, worker id: %s, task id: %s", res.GetRequestId(), workerId, res.GetTaskId())
			*ResponseChan <- res
			ResponseChanMap.Remove(res.GetRequestId())
		}(&s.ResponseChanMap, res)
	}
}
