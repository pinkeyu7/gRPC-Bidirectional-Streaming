package grpcstreaming

type errorInfo struct {
	Code    errorCode `json:"code"`
	Message string    `json:"message"`
}

type errorResponse struct {
	Error     *errorInfo `json:"error"`
	RequestID string     `json:"request_id"`
}

type errorCode uint64

const (
	ErrorCodeInternalServerError errorCode = 0
	ErrorCodeClientTimeout       errorCode = 1
	ErrorCodeNotFound            errorCode = 404
)

func NewErrorResponse[Response any](requestID string, errCode errorCode, errMessage string) *Response {
	errRes := errorResponse{
		Error: &errorInfo{
			Code:    errCode,
			Message: errMessage,
		},
		RequestID: requestID,
	}

	var res Response
	_ = convert(&errRes, &res)

	return &res
}
