package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type Handler func(ctx *Context)
type userCtxKey struct{}

const (
	AuthCookie = "auth-cookie"
)

type Context struct {
	context.Context
	r *http.Request
	w http.ResponseWriter

	closers []io.Closer
}

func (ctx *Context) ExtractUserIDFromContext() (string, error) {
	userID, ok := ctx.Value(userCtxKey{}).(string)
	if !ok {
		return "", fmt.Errorf("user id not found in context")
	}
	return userID, nil
}
