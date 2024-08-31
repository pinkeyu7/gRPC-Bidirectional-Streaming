package grpc_streaming

type errorInfo struct {
	Code    uint32 `json:"code"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error     *errorInfo `json:"error"`
	RequestId string     `json:"request_id"`
}
