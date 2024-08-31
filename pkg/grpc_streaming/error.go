package grpc_streaming

type errorInfo struct {
	Code    errorCode `json:"code"`
	Message string    `json:"message"`
}

type errorResponse struct {
	Error     *errorInfo `json:"error"`
	RequestId string     `json:"request_id"`
}

type errorCode uint32

const (
	ErrorCodeInternalServerError errorCode = 0
	ErrorCodeClientTimeout       errorCode = 1
	ErrorCodeNotFound            errorCode = 404
)

func NewErrorResponse[Response any](requestId string, errCode errorCode, errMessage string) *Response {
	errRes := errorResponse{
		Error: &errorInfo{
			Code:    errCode,
			Message: errMessage,
		},
		RequestId: requestId,
	}

	var res Response
	_ = convert(&errRes, &res)

	return &res
}
