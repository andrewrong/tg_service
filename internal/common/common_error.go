package common

import (
	"encoding/json"
)

type ErrorType string

var (
	ParameterError       ErrorType = "ParameterError"
	InitError            ErrorType = "InitError"
	ContextCancelError   ErrorType = "ContextCancelError"
	ExternalServiceError ErrorType = "ExternalServiceError"
)

type InnerError struct {
	ErrType ErrorType `json:"err_type"`
	ErrMsg  string    `json:"err_msg"`
	Code    int       `json:"code"`
}

func (e *InnerError) Error() string {
	data, _ := json.Marshal(e)
	return string(data)
}

func (e *InnerError) GetCode() int {
	return e.Code
}

func (e *InnerError) GetType() ErrorType {
	return e.ErrType
}
