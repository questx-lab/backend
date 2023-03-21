package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/utils/token"
)

type Handler func(ctx CustomContext)
type userCtxKey struct{}

type CustomContext struct {
	context.Context
	r *http.Request
	w http.ResponseWriter

	configs config.Configs
}

func UserIDToContext(ctx CustomContext) {
	prefix := "Bearer "
	authHeader := ctx.r.Header.Get("Authorization")
	reqToken := strings.TrimPrefix(authHeader, prefix)

	userID, err := token.Verify(reqToken, &ctx.configs)
	if err != nil {
		http.Error(ctx.w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx.Context = context.WithValue(ctx.Context, userCtxKey{}, userID)
}

func (ctx *CustomContext) ExtractUserIDFromContext() string {
	userID, ok := ctx.Value(userCtxKey{}).(string)
	if !ok {
		http.Error(ctx.w, fmt.Errorf("user id not found in context").Error(), http.StatusInternalServerError)
		return ""
	}
	return userID
}
