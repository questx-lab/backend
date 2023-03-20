package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/utils/token"
)

type CustomContext struct {
	r       *http.Request
	w       http.ResponseWriter
	Ctx     context.Context
	configs config.Configs
}

type userCtxKey struct{}

func (ctx *CustomContext) UserToContext() {
	prefix := "Bearer "
	authHeader := ctx.r.Header.Get("Authorization")
	reqToken := strings.TrimPrefix(authHeader, prefix)

	userID, err := token.Verify(reqToken, &ctx.configs)
	if err != nil {
		http.Error(ctx.w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx.Ctx = context.WithValue(ctx.Ctx, userCtxKey{}, userID)
}

func (ctx *CustomContext) ExtractUserIDFromContext() string {
	userID, ok := ctx.Ctx.Value(userCtxKey{}).(string)
	if !ok {
		http.Error(ctx.w, fmt.Errorf("user id not found in context").Error(), http.StatusInternalServerError)
		return ""
	}
	return userID
}
