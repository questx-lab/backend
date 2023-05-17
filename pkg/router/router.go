package router

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/pkg/authenticator"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/logger"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"gorm.io/gorm"
)

type HandlerFunc[Request, Response any] func(ctx xcontext.Context, req *Request) (*Response, error)
type MiddlewareFunc func(ctx xcontext.Context) error
type CloserFunc func(ctx xcontext.Context)
type WebsocketHandlerFunc[Request any] func(ctx xcontext.Context, req *Request) error

type Router struct {
	mux *http.ServeMux

	befores []MiddlewareFunc
	afters  []MiddlewareFunc
	closers []CloserFunc

	logger       logger.Logger
	cfg          config.Configs
	tokenEngine  authenticator.TokenEngine
	sessionStore sessions.Store
	httpClient   *http.Client
	db           *gorm.DB
}

func New(db *gorm.DB, cfg config.Configs, logger logger.Logger) *Router {
	r := &Router{
		mux:          http.NewServeMux(),
		cfg:          cfg,
		tokenEngine:  authenticator.NewTokenEngine(cfg.Auth.TokenSecret),
		sessionStore: sessions.NewCookieStore([]byte(cfg.Session.Secret)),
		logger:       logger,
		db:           db,
		httpClient:   &http.Client{},
	}

	r.AddCloser(handleResponse())
	return r
}

func GET[Request, Response any](router *Router, pattern string, handler HandlerFunc[Request, Response]) {
	route(router, http.MethodGet, pattern, handler)
}

func POST[Request, Response any](router *Router, pattern string, handler HandlerFunc[Request, Response]) {
	route(router, http.MethodPost, pattern, handler)
}

func Websocket[Request any](router *Router, pattern string, handler WebsocketHandlerFunc[Request]) {
	routeWS(router, pattern, handler)
}

func route[Request, Response any](router *Router, method, pattern string, handler HandlerFunc[Request, Response]) {
	befores := make([]MiddlewareFunc, len(router.befores))
	afters := make([]MiddlewareFunc, len(router.afters))
	closers := make([]CloserFunc, len(router.closers))

	copy(befores, router.befores)
	copy(afters, router.afters)
	copy(closers, router.closers)

	router.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		ctx := xcontext.NewContext(r.Context(), r, w, router.cfg, router.logger, router.db, nil)
		xcontext.SetHTTPClient(ctx, router.httpClient)

		var req Request
		err := parseRequest(ctx, method, &req)
		if err != nil {
			xcontext.SetError(ctx, err)
		}

		runMiddleware(ctx, befores, afters, closers, func() error {
			resp, err := handler(ctx, &req)
			if err != nil {
				return err
			} else if resp != nil {
				xcontext.SetResponse(ctx, resp)
			}
			return nil
		})
	})
}

func routeWS[Request any](router *Router, pattern string, handler WebsocketHandlerFunc[Request]) {
	befores := make([]MiddlewareFunc, len(router.befores))
	afters := make([]MiddlewareFunc, len(router.afters))
	closers := make([]CloserFunc, len(router.closers))

	copy(befores, router.befores)
	copy(afters, router.afters)
	copy(closers, router.closers)

	router.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			router.logger.Errorf("Cannot upgrade websocket: %v", err)
			return
		}

		ctx := xcontext.NewContext(r.Context(), r, w, router.cfg, router.logger, router.db, conn)
		xcontext.SetHTTPClient(ctx, router.httpClient)

		var req Request
		if err == nil {
			err = parseRequest(ctx, http.MethodGet, &req)
			if err != nil {
				xcontext.SetError(ctx, err)
				return
			}
		}

		runMiddleware(ctx, befores, afters, closers, func() error {
			if err := handler(ctx, &req); err != nil {
				return err
			}

			return nil
		})
	})
}

func (r *Router) Before(middleware MiddlewareFunc) {
	r.befores = append(r.befores, middleware)
}

func (r *Router) After(middleware MiddlewareFunc) {
	r.afters = append(r.afters, middleware)
}

func (r *Router) AddCloser(closer CloserFunc) {
	r.closers = append(r.closers, closer)
}

func (r *Router) Static(root, relativePath string) {
	r.mux.Handle(root, http.FileServer(http.Dir(relativePath)))
}

func (r *Router) Branch() *Router {
	clone := *r
	copy(clone.befores, r.befores)
	copy(clone.afters, r.afters)
	copy(clone.closers, r.closers)
	return &clone
}

func (r *Router) Handler() http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins:     []string{"http://localhost:3000", "http://35.247.96.16"},
		AllowCredentials:   true,
		Debug:              true,
		OptionsPassthrough: true,
		AllowedHeaders:     []string{"Origin", "Access-Control-Allow-Origin", "X-Requested-With", "Content-Type", "Accept", "Authorization"},
	})
	return c.Handler(r.mux)
}

func parseBody(r *http.Request, req any) error {
	switch r.Method {
	case http.MethodGet:
		v := reflect.ValueOf(req).Elem()
		for i := 0; i < v.NumField(); i++ {
			name := v.Type().Field(i).Tag.Get("json")
			queryVal := r.URL.Query().Get(name)
			if queryVal == "" {
				continue
			}

			pointer := v.Field(i).Addr().Interface()

			switch v.Field(i).Kind() {
			case reflect.String:
				p := pointer.(*string)
				*p = queryVal

			case reflect.Int:
				p := pointer.(*int)
				val, err := strconv.Atoi(queryVal)
				if err != nil {
					return err
				}

				*p = val
			case reflect.Bool:
				p := pointer.(*bool)
				val, err := strconv.ParseBool(queryVal)
				if err != nil {
					return err
				}

				*p = val
			}
		}

	case http.MethodPost:
		if r.Header.Get("Content-type") == "application/json" {
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return err
			}

			if err := json.Unmarshal(b, &req); err != nil {
				return err
			}
		}

	default:
		return errors.New("unsupported method")
	}

	return nil
}

func parseSession(ctx xcontext.Context, req any) error {
	session, err := ctx.SessionStore().Get(ctx.Request(), ctx.Configs().Session.Name)
	if err != nil {
		ctx.Logger().Errorf("Cannot decode the existing session: %v", err)
		return nil
	}

	v := reflect.ValueOf(req).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		tagValue := v.Type().Field(i).Tag.Get("session")
		name, action, found := strings.Cut(tagValue, ",")
		if name == "" {
			continue
		}

		value, ok := session.Values[name]
		if !ok {
			return errors.New("session value not found")
		}

		if found {
			if action == "delete" {
				delete(session.Values, name)
			} else {
				return errors.New("invalid session tag")
			}
		}

		if reflect.TypeOf(value) != field.Type() {
			return errors.New("invalid value type in session")
		}

		if field.CanSet() {
			field.Set(reflect.ValueOf(value))
		}
	}

	if err := session.Save(ctx.Request(), ctx.Writer()); err != nil {
		return err
	}

	return nil
}

func parseRequest(ctx xcontext.Context, method string, req any) error {
	if method != ctx.Request().Method {
		return errorx.New(errorx.BadRequest, "Not supported method %s", ctx.Request().Method)
	}

	if err := parseBody(ctx.Request(), req); err != nil {
		ctx.Logger().Errorf("Cannot bind the body: %v", err)
		return errorx.Unknown
	}

	if err := parseSession(ctx, req); err != nil {
		ctx.Logger().Errorf("Cannot find the session: %v", err)
		return errorx.New(errorx.BadRequest, "Cannot find the session")
	}

	return nil
}

func runMiddleware(
	ctx xcontext.Context,
	befores, afters []MiddlewareFunc,
	closers []CloserFunc,
	handler func() error,
) {
	func() {
		if xcontext.GetError(ctx) != nil {
			return
		}

		for _, m := range befores {
			if err := m(ctx); err != nil {
				xcontext.SetError(ctx, err)
				return
			}
		}

		if err := handler(); err != nil {
			xcontext.SetError(ctx, err)
			return
		}

		for _, m := range afters {
			if err := m(ctx); err != nil {
				xcontext.SetError(ctx, err)
				return
			}
		}
	}()

	for _, closer := range closers {
		closer(ctx)
	}
}
