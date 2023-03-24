package errorx

import "fmt"

// Error is the only error type that is able to send to client. Other error
// types will be filtered by the router.
type Error struct {
	Code    uint64
	Message string
}

func (e Error) Error() string {
	return e.Message
}

// New returns a wrapper error of target. The internal will be included in the
// new error but it is never displayed at the client side.
func New(internal error, clientFacing Error) error {
	return fmt.Errorf("%v: %w", internal, clientFacing)
}

// NewGeneric returns an ErrGeneric object with a customized message. If the
// internal error is specified, the function will return a wrapper error
// instead. See New.
//
// ErrGeneric should be the only Error which allowed to customize message. DO
// NOT customize other Error objects.
func NewGeneric(internal error, clientFacingFormat string, args ...any) error {
	errx := Error{
		Code:    ErrGeneric.Code,
		Message: fmt.Sprintf(clientFacingFormat, args...),
	}

	if internal == nil {
		return errx
	}

	return New(internal, errx)
}
