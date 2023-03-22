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

	middlewares       []gin.HandlerFunc
	cfg               config.Configs
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
	r.Inner.GET(pattern, append(r.middlewares, wrapHandler(r, "GET", handler))...)
}

func POST[Request, Response any](r *Router, pattern string, handler HandlerFunc[Request, Response]) {
	r.Inner.POST(pattern, append(r.middlewares, wrapHandler(r, "POST", handler))...)
}

func (r *Router) Use(middleware MiddlewareFunc) {
	r.middlewares = append(r.middlewares, wrapMiddleware(r, middleware))
}

func (r *Router) Group(pattern string) *Router {
	group := r.Branch()
	group.Inner = r.Inner.Group(pattern)
	return group
}

func (r *Router) Branch() *Router {
	clone := &Router{
		Inner:             r.Inner,
		cfg:               r.cfg,
		accessTokenEngine: r.accessTokenEngine,
	}
	copy(clone.middlewares, r.middlewares)

	return clone
}

func (r *Router) Static(relativePath, root string) {
	r.Inner.Static(relativePath, root)
}

func (r *Router) Handler() http.Handler {
	return r.Inner.(*gin.Engine)
}
