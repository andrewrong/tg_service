package common

import (
	"context"
	"net/url"
)

type Checker interface {
	// Check checks the internal state of the object and returns an error if any.
	//
	// It returns a pointer to an InnerError object, which represents an error that
	// occurred during the check. If the check was successful, it returns nil.
	//
	// Example:
	//   err := obj.Check()
	//   if err != nil {
	//       // handle the error
	//   }
	Check() error
}

type AI interface {
	GetCompletion(prompt, systemMessage, mode string, temperature float32, jsonModel bool, ctx context.Context) (string, error)
}

// isValidURL 解析并验证 URL 的合法性
func IsValidURL(u string) (*url.URL, bool) {
	if u == "" {
		return nil, false
	}
	parsedURL, err := url.ParseRequestURI(u)
	if err != nil {
		return nil, false
	}
	return parsedURL, true
}
