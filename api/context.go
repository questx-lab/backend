package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/utils/token"
)

type Handler func(ctx Context)
type userCtxKey struct{}

const (
	AuthCookie = "auth-cookie"
)

type Context struct {
	context.Context
	r *http.Request
	w http.ResponseWriter

	configs config.Configs
}

func UserIDToContext(ctx Context) {
	reqToken, err := ctx.r.Cookie(AuthCookie)
	if err != nil {
		http.Error(ctx.w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, err := token.Verify(reqToken.Value, &ctx.configs)
	if err != nil {
		http.Error(ctx.w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx.Context = context.WithValue(ctx.Context, userCtxKey{}, userID)
}

func (ctx *Context) ExtractUserIDFromContext() string {
	userID, ok := ctx.Value(userCtxKey{}).(string)
	if !ok {
		http.Error(ctx.w, fmt.Errorf("user id not found in context").Error(), http.StatusInternalServerError)
		return ""
	}
	return userID
}
