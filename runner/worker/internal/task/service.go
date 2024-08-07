package task

import taskProto "grpc-bidirectional-streaming/pb/task"

type Service interface {
	HandleRequest(req *taskProto.RequestFromServerRequest) (*taskProto.RequestFromServerResponse, error)
}
