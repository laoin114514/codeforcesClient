package codeforcesClient

import "fmt"

type ErrorCode int

const (
	ErrNetwork     ErrorCode = iota
	ErrAPI
	ErrRateLimit
	ErrAuth
	ErrInvalidParam
)

type CFError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e *CFError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *CFError) Unwrap() error {
	return e.Cause
}

var _ error = (*CFError)(nil)
