package errorx

type Error struct {
	Code    uint64
	Message string
}

func (e Error) Error() string {
	return e.Message
}
