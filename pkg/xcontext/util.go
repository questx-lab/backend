package xcontext

import (
	"context"
	"net/http"
)

type (
	userIDKey     struct{}
	responseKey   struct{}
	errorKey      struct{}
	httpClientKey struct{}
)

func SetError(ctx Context, err error) {
	ctx.Set(errorKey{}, err)
}

func GetError(ctx context.Context) error {
	err := ctx.Value(errorKey{})
	if err == nil {
		return nil
	}

	return err.(error)
}

func SetResponse(ctx Context, resp any) {
	ctx.Set(responseKey{}, resp)
}

func GetResponse(ctx context.Context) any {
	return ctx.Value(responseKey{})
}

func SetRequestUserID(ctx Context, id string) {
	ctx.Set(userIDKey{}, id)
}

func GetRequestUserID(ctx context.Context) string {
	id := ctx.Value(userIDKey{})
	if id == nil {
		return ""
	}

	return id.(string)
}

func SetHTTPClient(ctx Context, client *http.Client) {
	ctx.Set(httpClientKey{}, client)
}

func GetHTTPClient(ctx context.Context) *http.Client {
	client := ctx.Value(httpClientKey{})
	if client == nil {
		return http.DefaultClient
	}

	return client.(*http.Client)
}
