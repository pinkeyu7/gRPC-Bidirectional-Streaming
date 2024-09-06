package grpcstreaming

var clientID string

func SetClientID(cID string) {
	clientID = cID
}

type clientObject[Request any, Response any] interface {
	Send(*Response) error
	Recv() (*Request, error)
}
