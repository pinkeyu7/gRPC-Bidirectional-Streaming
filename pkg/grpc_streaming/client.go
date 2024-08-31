package grpc_streaming

var clientId string

func SetClientId(cId string) {
	clientId = cId
}

type clientObject[Request any, Response any] interface {
	Send(*Response) error
	Recv() (*Request, error)
}
