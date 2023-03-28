package errorx

type Code int

var Unknown = Error{Code: 100000, Message: "Request failed"}

const (
	// Common codes
	BadRequest       Code = 100001
	BadResponse      Code = 100002
	PermissionDenied Code = 100003
	NotFound         Code = 100004
	Unauthenticated  Code = 100005
	AlreadyExists    Code = 100006
	Internal         Code = 100007
	Unavailable      Code = 100008
)
