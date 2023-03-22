package api

import (
	"context"
	"net/http"
	"strings"

	"io"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/jwt"
	"github.com/questx-lab/backend/pkg/session"
)

type Handler func(ctx *Context) error

type Context struct {
	context.Context

	r       *http.Request
	w       http.ResponseWriter
	closers []io.Closer

	AccessTokenEngine *jwt.Engine[model.AccessToken]
	SessionStore      *session.Store

	Configs config.Configs
}

func (ctx *Context) GetRequest() *http.Request {
	return ctx.r
}

func (ctx *Context) GetResponse() http.ResponseWriter {
	return ctx.w
}

func (ctx *Context) GetUserID() string {
	if token := ctx.getAccessToken(); token != "" {
		if info, err := ctx.AccessTokenEngine.Verify(token); err == nil {
			return info.ID
		}
	}

	return ""
}

func (ctx *Context) getAccessToken() string {
	authorization := ctx.r.Header.Get("Authorization")
	auth, token, found := strings.Cut(authorization, " ")
	if found {
		if auth == "Bearer" {
			return token
		}
		return ""
	}

	cookie, err := ctx.r.Cookie(ctx.Configs.Auth.AccessTokenName)
	if err != nil {
		return ""
	}

	return cookie.Value
}
