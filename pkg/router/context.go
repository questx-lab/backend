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
	"gorm.io/gorm"
)

type (
	userIDKey   struct{}
	responseKey struct{}
	errorKey    struct{}
)

type Context interface {
	// This Context is an extension of the context.Context. It's compatible with any library using
	// the context.Context.
	context.Context

	// Request returns the *http.Request.
	Request() *http.Request

	// Writer returns the http.ResponseWriter.
	Writer() http.ResponseWriter

	// Set is another implementation of context.WithValue.
	Set(key, value any)

	// Get returns the Value corresponding to the given key which is stored by Set method or
	// context.WithValue function.
	Get(key any) any

	// GetUserID returns the user id corresponding to the authentication cookie or authorization
	// header. After the first parsing, the user id should be stored into the Context for future
	// usage.
	GetUserID() string

	// SetResponse sets the response object sent to client. This method would be used in the
	// Router.
	SetResponse(resp any)

	// GetResponse gets the response object sent to client. This method only can return non-nil
	// object in After middlewares.
	GetResponse() any

	// SessionStore returns the sessions.Store corresponding to this request.
	SessionStore() sessions.Store

	// AccessTokenEngine returns the TokenEngine for model.AccessToken struct.
	AccessTokenEngine() authenticator.TokenEngine[model.AccessToken]

	// Configs returns the configurations.
	Configs() config.Configs

	// SetError sets the Error into Context. This method should be used in only Router.
	SetError(err error)

	// Error returns the error that is set by SetError method.
	Error() error

	// Logger returns the logger.
	Logger() logger.Logger

	// DB returns the gorm.DB.
	DB() *gorm.DB

	// BeginTx replaces the returned value of DB() method by a database transaction.
	BeginTx()

	// CommitTx commits the transaction if it exists.
	CommitTx()

	// RollbackTx rollbacks the transaction if it exists.
	RollbackTx()
}

type defaultContext struct {
	context.Context

	r *http.Request
	w http.ResponseWriter

	accessTokenEngine authenticator.TokenEngine[model.AccessToken]
	sessionStore      sessions.Store
	configs           config.Configs
	logger            logger.Logger

	db *gorm.DB
	tx *gorm.DB
}

func NewContext(
	ctx context.Context,
	r *http.Request,
	w http.ResponseWriter,
	cfg config.Configs,
	logger logger.Logger,
	db *gorm.DB,
) *defaultContext {
	return &defaultContext{
		Context: ctx,
		r:       r, w: w,
		accessTokenEngine: authenticator.NewTokenEngine[model.AccessToken](cfg.Token),
		sessionStore:      sessions.NewCookieStore([]byte(cfg.Session.Secret)),
		configs:           cfg,
		logger:            logger,
		db:                db,
		tx:                nil,
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

func (ctx *defaultContext) SetError(err error) {
	ctx.Context = context.WithValue(ctx.Context, errorKey{}, err)
}

func (ctx *defaultContext) Error() error {
	err := ctx.Value(errorKey{})
	if err != nil {
		return err.(error)
	}
	return nil
}

func (ctx *defaultContext) Logger() logger.Logger {
	return ctx.logger
}

func (ctx *defaultContext) DB() *gorm.DB {
	if ctx.tx != nil {
		return ctx.tx
	}
	return ctx.db
}

func (ctx *defaultContext) BeginTx() {
	ctx.tx = ctx.db.Begin()
}

func (ctx *defaultContext) CommitTx() {
	if ctx.tx != nil {
		ctx.tx.Commit()
		ctx.tx = nil
	}
}

func (ctx *defaultContext) RollbackTx() {
	if ctx.tx != nil {
		ctx.tx.Rollback()
		ctx.tx = nil
	}
}
