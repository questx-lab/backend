package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/jwt"
)

type Context struct {
	context.Context

	Request *http.Request
	Writer  http.ResponseWriter

	cfg config.Configs
}

func (ctx Context) GetUserID() string {
	verifier := jwt.NewVerifier[model.AccessToken](ctx.cfg.Auth.TokenSecret)
	if token := ctx.getAccessToken(); token != "" {
		if info, err := verifier.Verify(token); err == nil {
			return info.ID
		}
	}

	return ""
}

func (ctx Context) getAccessToken() string {
	authorization := ctx.Request.Header.Get("Authorization")
	auth, token, found := strings.Cut(authorization, " ")
	if found {
		if auth == "Bearer" {
			return token
		}
		return ""
	}

	cookie, err := ctx.Request.Cookie(ctx.cfg.Auth.AccessTokenName)
	if err != nil {
		return ""
	}

	return cookie.Value
}
