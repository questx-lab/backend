package router

import (
	"context"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/authenticator"
	"github.com/questx-lab/backend/pkg/logger"
)

type (
	userIDKey   struct{}
	responseKey struct{}
)

type Context interface {
	context.Context

	Request() *http.Request
	SetRequest(*http.Request)
	Writer() http.ResponseWriter

	Set(key, value any)
	Get(key any) any
	GetUserID() string

	SetResponse(resp any)
	GetResponse() any
	OverrideResponse()

	SessionStore() sessions.Store
	AccessTokenEngine() authenticator.TokenEngine[model.AccessToken]
	SetAccessTokenEngine(authenticator.TokenEngine[model.AccessToken])

	Configs() config.Configs

	SetError(err error)
	Error() error

	Logger() logger.Logger
}

type defaultContext struct {
	context.Context

	err error
	r   *http.Request
	w   http.ResponseWriter

	accessTokenEngine authenticator.TokenEngine[model.AccessToken]
	sessionStore      sessions.Store
	configs           config.Configs

	logger logger.Logger
}

func NewContext(ctx context.Context, r *http.Request, w http.ResponseWriter, cfg config.Configs) *defaultContext {
	return &defaultContext{
		Context: ctx,
		r:       r, w: w,
		accessTokenEngine: authenticator.NewTokenEngine[model.AccessToken](cfg.Token),
		sessionStore:      sessions.NewCookieStore([]byte(cfg.Session.Secret)),
		logger:            logger.NewLogger(),
		configs:           cfg,
	}
}

func (ctx *defaultContext) GetUserID() string {
	if value := ctx.Get(userIDKey{}); value != nil {
		return value.(string)
	}

	if token := ctx.getAccessToken(); token != "" {
		if info, err := ctx.accessTokenEngine.Verify(token); err == nil {
			ctx.Set(userIDKey{}, info.ID)
			return info.ID
		}
	}

	return ""
}

func (ctx *defaultContext) Set(key, value any) {
	ctx.Context = context.WithValue(ctx.Context, key, value)
}

func (ctx *defaultContext) Get(key any) any {
	return ctx.Context.Value(key)
}

func (ctx *defaultContext) SetResponse(resp any) {
	ctx.Set(responseKey{}, resp)
}

func (ctx *defaultContext) GetResponse() any {
	return ctx.Get(responseKey{})
}

func (ctx *defaultContext) OverrideResponse() {
	ctx.Set(responseKey{}, nil)
}

func (ctx *defaultContext) getAccessToken() string {
	authorization := ctx.r.Header.Get("Authorization")
	auth, token, found := strings.Cut(authorization, " ")
	if found {
		if auth == "Bearer" {
			return token
		}
		return ""
	}

	cookie, err := ctx.r.Cookie(ctx.configs.Auth.AccessTokenName)
	if err != nil {
		return ""
	}

	return cookie.Value
}

func (ctx *defaultContext) Request() *http.Request {
	return ctx.r
}

func (ctx *defaultContext) Writer() http.ResponseWriter {
	return ctx.w
}

func (ctx *defaultContext) AccessTokenEngine() authenticator.TokenEngine[model.AccessToken] {
	return ctx.accessTokenEngine
}

func (ctx *defaultContext) SessionStore() sessions.Store {
	return ctx.sessionStore
}

func (ctx *defaultContext) Configs() config.Configs {
	return ctx.configs
}

func (ctx *defaultContext) SetRequest(r *http.Request) {
	ctx.r = r
}

func (ctx *defaultContext) SetAccessTokenEngine(a authenticator.TokenEngine[model.AccessToken]) {
	ctx.accessTokenEngine = a
}

func DefaultContext() Context {
	return &defaultContext{
		Context: context.Background(),
	}
}

func (ctx *defaultContext) SetError(err error) {
	ctx.err = err
}

func (ctx *defaultContext) Error() error {
	return ctx.err
}

func (ctx *defaultContext) Logger() logger.Logger {
	return ctx.logger
}
