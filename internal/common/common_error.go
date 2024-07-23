package common

import "fmt"

type ErrorType string

var (
	ParameterError       ErrorType = "ParameterError"
	InitError            ErrorType = "InitError"
	ContextCancelError   ErrorType = "ContextCancelError"
	ExternalServiceError ErrorType = "ExternalServiceError"
)

type InnerError struct {
	ErrType ErrorType
	ErrMsg  string
	Code    int
}

func (e *InnerError) Error() string {
	return fmt.Sprintf("type: %s, msg: %s, code:%d", e.ErrType, e.ErrMsg, e.Code)
}

func (e *InnerError) GetCode() int {
	return e.Code
}

func (e *InnerError) GetType() ErrorType {
	return e.ErrType
}
