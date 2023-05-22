package xcontext

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/pkg/authenticator"
	"github.com/questx-lab/backend/pkg/logger"
	"github.com/questx-lab/backend/pkg/ws"
	"gorm.io/gorm"
)

type (
	userIDKey       struct{}
	responseKey     struct{}
	errorKey        struct{}
	httpClientKey   struct{}
	httpRequestKey  struct{}
	httpWriterKey   struct{}
	sessionStoreKey struct{}
	tokenEngineKey  struct{}
	configsKey      struct{}
	loggerKey       struct{}
	wsClientKey     struct{}
	dbKey           struct{}
	dbTxKey         struct{}
)

func WithError(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, errorKey{}, err)
}

func Error(ctx context.Context) error {
	err := ctx.Value(errorKey{})
	if err == nil {
		return nil
	}

	return err.(error)
}

func WithResponse(ctx context.Context, resp any) context.Context {
	return context.WithValue(ctx, responseKey{}, resp)
}

func Response(ctx context.Context) any {
	return ctx.Value(responseKey{})
}

func WithRequestUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userIDKey{}, id)
}

func RequestUserID(ctx context.Context) string {
	id := ctx.Value(userIDKey{})
	if id == nil {
		return ""
	}

	return id.(string)
}

func WithHTTPClient(ctx context.Context, client *http.Client) context.Context {
	return context.WithValue(ctx, httpClientKey{}, client)
}

func HTTPClient(ctx context.Context) *http.Client {
	client := ctx.Value(httpClientKey{})
	if client == nil {
		return http.DefaultClient
	}

	return client.(*http.Client)
}

func WithHTTPRequest(ctx context.Context, req *http.Request) context.Context {
	return context.WithValue(ctx, httpRequestKey{}, req)
}

func HTTPRequest(ctx context.Context) *http.Request {
	req := ctx.Value(httpRequestKey{})
	if req == nil {
		return nil
	}

	return req.(*http.Request)
}

func WithHTTPWriter(ctx context.Context, writer http.ResponseWriter) context.Context {
	return context.WithValue(ctx, httpWriterKey{}, writer)
}

func HTTPWriter(ctx context.Context) http.ResponseWriter {
	writer := ctx.Value(httpWriterKey{})
	if writer == nil {
		return nil
	}

	return writer.(http.ResponseWriter)
}

func WithSessionStore(ctx context.Context, store sessions.Store) context.Context {
	return context.WithValue(ctx, sessionStoreKey{}, store)
}

func SessionStore(ctx context.Context) sessions.Store {
	store := ctx.Value(sessionStoreKey{})
	if store == nil {
		return nil
	}

	return store.(sessions.Store)
}

func WithTokenEngine(ctx context.Context, engine authenticator.TokenEngine) context.Context {
	return context.WithValue(ctx, tokenEngineKey{}, engine)
}

func TokenEngine(ctx context.Context) authenticator.TokenEngine {
	engine := ctx.Value(tokenEngineKey{})
	if engine == nil {
		return nil
	}

	return engine.(authenticator.TokenEngine)
}

func WithConfigs(ctx context.Context, cfg config.Configs) context.Context {
	return context.WithValue(ctx, configsKey{}, cfg)
}

func Configs(ctx context.Context) config.Configs {
	cfg := ctx.Value(configsKey{})
	if cfg == nil {
		return config.Configs{}
	}

	return cfg.(config.Configs)
}

func WithLogger(ctx context.Context, logger logger.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

func Logger(ctx context.Context) logger.Logger {
	lg := ctx.Value(loggerKey{})
	if lg == nil {
		return nil
	}

	return lg.(logger.Logger)
}

func WithWSClient(ctx context.Context, client *ws.Client) context.Context {
	return context.WithValue(ctx, wsClientKey{}, client)
}

func WSClient(ctx context.Context) *ws.Client {
	client := ctx.Value(wsClientKey{})
	if client == nil {
		return nil
	}

	return client.(*ws.Client)
}

func WithDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, dbKey{}, db)
}

func DB(ctx context.Context) *gorm.DB {
	if tx := DBTransaction(ctx); tx != nil {
		return tx
	}

	db := ctx.Value(dbKey{})
	if db == nil {
		return nil
	}

	return db.(*gorm.DB)
}

func WithDBTransaction(ctx context.Context) context.Context {
	return context.WithValue(ctx, dbTxKey{}, DB(ctx).Begin())
}

func DBTransaction(ctx context.Context) *gorm.DB {
	tx := ctx.Value(dbTxKey{})
	if tx == nil {
		return nil
	}

	return tx.(*gorm.DB)
}

func WithCommitDBTransaction(ctx context.Context) context.Context {
	if tx := DBTransaction(ctx); tx != nil {
		tx.Commit()
		return context.WithValue(ctx, dbTxKey{}, nil)
	}

	return ctx
}

func WithRollbackDBTransaction(ctx context.Context) context.Context {
	if tx := DBTransaction(ctx); tx != nil {
		tx.Rollback()
		return context.WithValue(ctx, dbTxKey{}, nil)
	}

	return ctx
}

// set websocket connection  for context
func (ctx *defaultContext) SetWsConn(conn *websocket.Conn) {
	ctx.websocket = conn
}

// returns the websocket connection of a request
func (ctx *defaultContext) WsConn() *websocket.Conn {
	return ctx.websocket
}
