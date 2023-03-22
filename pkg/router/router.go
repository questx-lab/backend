package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/jwt"
)

type HandlerFunc[Request, Response any] func(ctx *Context, req Request) (*Response, error)
type MiddlewareFunc func(ctx *Context) error

type Router struct {
	Inner gin.IRouter
	cfg   config.Configs

	accessTokenEngine *jwt.Engine[model.AccessToken]
}

func New(cfg config.Configs) *Router {
	return &Router{
		Inner:             gin.New(),
		accessTokenEngine: jwt.NewEngine[model.AccessToken](cfg.Token),
		cfg:               cfg,
	}
}

func GET[Request, Response any](r *Router, pattern string, handler HandlerFunc[Request, Response]) {
	r.Inner.GET(pattern, wrapHandler(r, "GET", handler))
}

func POST[Request, Response any](r *Router, pattern string, handler HandlerFunc[Request, Response]) {
	r.Inner.POST(pattern, wrapHandler(r, "POST", handler))
}

func (r *Router) Use(middleware MiddlewareFunc) {
	r.Inner.Use(wrapMiddleware(r, middleware))
}

func (r *Router) Group(pattern string) *Router {
	return &Router{
		Inner:             r.Inner.Group(pattern),
		cfg:               r.cfg,
		accessTokenEngine: r.accessTokenEngine,
	}
}

func (r *Router) Static(relativePath, root string) {
	r.Inner.Static(relativePath, root)
}

func (r *Router) Handler() http.Handler {
	return r.Inner.(*gin.Engine)
}
