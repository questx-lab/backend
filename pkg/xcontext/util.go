package xcontext

type (
	userIDKey   struct{}
	responseKey struct{}
	errorKey    struct{}
)

func SetError(ctx Context, err error) {
	ctx.Set(errorKey{}, err)
}

func GetError(ctx Context) error {
	err := ctx.Get(errorKey{})
	if err == nil {
		return nil
	}

	return err.(error)
}

func SetResponse(ctx Context, resp any) {
	ctx.Set(responseKey{}, resp)
}

func GetResponse(ctx Context) any {
	return ctx.Get(responseKey{})
}

func SetRequestUserID(ctx Context, id string) {
	ctx.Set(userIDKey{}, id)
}

func GetRequestUserID(ctx Context) string {
	id := ctx.Get(userIDKey{})
	if id == nil {
		return ""
	}

	return id.(string)
}
