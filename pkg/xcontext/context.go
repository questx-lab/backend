package xcontext

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/pkg/authenticator"
	"github.com/questx-lab/backend/pkg/logger"
	"gorm.io/gorm"
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

	// SessionStore returns the sessions.Store corresponding to this request.
	SessionStore() sessions.Store

	// TokenEngine supports to generate and verify token.
	TokenEngine() authenticator.TokenEngine

	// Configs returns the configurations.
	Configs() config.Configs

	// Logger returns the logger.
	Logger() logger.Logger

	// DB returns the gorm.DB.
	DB() *gorm.DB

	// BeginTx replaces the returned value of DB() method by a database transaction.
	BeginTx()

	// CommitTx commits the transaction.
	CommitTx()

	// RollbackTx rollbacks the transaction.
	RollbackTx()
}

type defaultContext struct {
	context.Context

	r *http.Request
	w http.ResponseWriter

	tokenEngine  authenticator.TokenEngine
	sessionStore sessions.Store
	configs      config.Configs
	logger       logger.Logger

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
		tokenEngine:  authenticator.NewTokenEngine(cfg.Auth.TokenSecret),
		sessionStore: sessions.NewCookieStore([]byte(cfg.Session.Secret)),
		configs:      cfg,
		logger:       logger,
		db:           db,
		tx:           nil,
	}
}

func (ctx *defaultContext) Set(key, value any) {
	ctx.Context = context.WithValue(ctx.Context, key, value)
}

func (ctx *defaultContext) Get(key any) any {
	return ctx.Context.Value(key)
}

func (ctx *defaultContext) Request() *http.Request {
	return ctx.r
}

func (ctx *defaultContext) Writer() http.ResponseWriter {
	return ctx.w
}

func (ctx *defaultContext) TokenEngine() authenticator.TokenEngine {
	return ctx.tokenEngine
}

func (ctx *defaultContext) SessionStore() sessions.Store {
	return ctx.sessionStore
}

func (ctx *defaultContext) Configs() config.Configs {
	return ctx.configs
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
