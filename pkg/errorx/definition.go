package errorx

var (
	ErrGeneric            = Error{1000, "generic error"}
	ErrBadRequest         = Error{1001, "bad request"}
	ErrNotSupportedMethod = Error{1002, "not supported method"}
	ErrBadResponse        = Error{1003, "bad response"}
)
