package grpc_streaming

type Error struct {
	Code    uint32 `json:"code"`
	Message string `json:"message"`
}
