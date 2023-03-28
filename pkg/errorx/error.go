package errorx

import "fmt"

// Error is the only error type that is able to send to client. Other error
// types will be filtered by the router.
type Error struct {
	Code    Code
	Message string
}

func (e Error) Error() string {
	return e.Message
}

func New(code Code, clientFacingFormat string, args ...any) error {
	return Error{
		Code:    code,
		Message: fmt.Sprintf(clientFacingFormat, args...),
	}
}
