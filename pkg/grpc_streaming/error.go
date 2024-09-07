package grpcstreaming

type ErrorInfo struct {
	Code    errorCode `json:"code"`
	Message string    `json:"message"`
}

type errorResponse struct {
	Error     *ErrorInfo `json:"error"`
	RequestID string     `json:"request_id"`
}

type errorCode uint64

const (
	ErrorCodeUnknownError        errorCode = 500000
	ErrorCodeSendRequest         errorCode = 500001
	ErrorCodeSendResponse        errorCode = 500002
	ErrorCodeSetField            errorCode = 500003
	ErrorCodeRequestChanNotFound errorCode = 500004
	ErrorCodeRequestMarshal      errorCode = 500005
	ErrorCodeConvertStruct       errorCode = 500006
	ErrorCodeRetrieveError       errorCode = 500007
	ErrorCodeReplyMarshal        errorCode = 500008
	ErrorCodeBadRequest          errorCode = 400000
	ErrorCodeClientTimeout       errorCode = 400001
	ErrorCodeServerTimeout       errorCode = 400002
	ErrorCodeUnauthorized        errorCode = 400401
	ErrorCodeForbidden           errorCode = 400403
	ErrorCodeNotFound            errorCode = 400404
)

func NewError(code errorCode, message string) *ErrorInfo {
	return &ErrorInfo{
		Code:    code,
		Message: message,
	}
}

func NewErrorResponse[Response any](requestID string, errCode errorCode, message string) *Response {
	errRes := errorResponse{
		Error:     NewError(errCode, message),
		RequestID: requestID,
	}

	var res Response
	_ = convert(&errRes, &res)

	return &res
}
