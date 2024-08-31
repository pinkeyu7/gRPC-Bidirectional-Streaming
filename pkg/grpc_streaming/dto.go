package grpc_streaming

type Error struct {
	Code    uint32 `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error     *Error `json:"error"`
	RequestId string `json:"request_id"`
}
