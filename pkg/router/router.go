package router

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/ws"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

type HandlerFunc[Request, Response any] func(ctx context.Context, req *Request) (*Response, error)
type MiddlewareFunc func(ctx context.Context) (context.Context, error)
type CloserFunc func(ctx context.Context)
type WebsocketHandlerFunc[Request any] func(ctx context.Context, req *Request) error

type Router struct {
	mux *http.ServeMux
	ctx context.Context

	befores []MiddlewareFunc
	afters  []MiddlewareFunc
	closers []CloserFunc
}

func New(ctx context.Context) *Router {
	r := &Router{
		mux: http.NewServeMux(),
		ctx: ctx,
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
		ctx := router.ctx
		ctx = xcontext.WithHTTPRequest(ctx, r)
		ctx = xcontext.WithHTTPWriter(ctx, w)

		var req Request
		err := parseRequest(ctx, pattern, method, &req)
		if err != nil {
			ctx = xcontext.WithError(ctx, err)
		}

		runMiddleware(ctx, befores, afters, closers, func(ctx context.Context) (any, error) {
			return handler(ctx, &req)
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
			xcontext.Logger(router.ctx).Errorf("Cannot upgrade websocket: %v", err)
			return
		}

		ctx := router.ctx
		ctx = xcontext.WithHTTPRequest(ctx, r)
		ctx = xcontext.WithHTTPWriter(ctx, w)
		ctx = xcontext.WithWSClient(ctx, ws.NewClient(conn))

		var req Request
		if err == nil {
			err = parseRequest(ctx, pattern, http.MethodGet, &req)
			if err != nil {
				ctx = xcontext.WithError(ctx, err)
			}
		}

		runMiddleware(ctx, befores, afters, closers, func(ctx context.Context) (any, error) {
			if err := handler(ctx, &req); err != nil {
				return nil, err
			}

			return nil, nil
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

func (r *Router) Handler(cfg config.ServerConfigs) http.Handler {
	return cors.New(cors.Options{
		AllowedOrigins: cfg.AllowCORS,
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
		},
		AllowedHeaders:     []string{"*"},
		AllowCredentials:   true,
		OptionsPassthrough: true,
	}).Handler(r.mux)
}

func parseBody(ctx context.Context, r *http.Request, req any) error {
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

			case reflect.Int64:
				p := pointer.(*int64)
				val, err := strconv.ParseInt(queryVal, 10, 64)
				if err != nil {
					return err
				}

				*p = val

			case reflect.Uint64:
				p := pointer.(*uint64)
				val, err := strconv.ParseUint(queryVal, 10, 64)
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

			case reflect.Struct:

			default:
				return fmt.Errorf("not setting up for type %s", v.Field(i).Kind())
			}
		}

	case http.MethodPost:
		if r.Header.Get("Content-type") == "application/json" {
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return err
			}

			if err := json.Unmarshal(b, &req); err != nil {
				xcontext.Logger(ctx).Warnf("Got an invalid request: %v", b)
				return err
			}
		}

	default:
		return errors.New("unsupported method")
	}

	return nil
}

func parseSession(ctx context.Context, req any) error {
	httpRequest := xcontext.HTTPRequest(ctx)
	session, err := xcontext.SessionStore(ctx).Get(httpRequest, xcontext.Configs(ctx).Session.Name)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Cannot decode the existing session: %v", err)
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

	if err := session.Save(httpRequest, xcontext.HTTPWriter(ctx)); err != nil {
		return err
	}

	return nil
}

func parseRequest(ctx context.Context, pattern, method string, req any) error {
	httpRequest := xcontext.HTTPRequest(ctx)
	if pattern != httpRequest.URL.Path {
		return errorx.New(errorx.NotFound, "Not found api")
	}

	if method != httpRequest.Method {
		return errorx.New(errorx.NotFound, "Not supported method %s", httpRequest.Method)
	}

	if err := parseBody(ctx, httpRequest, req); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot bind the body: %v", err)
		return errorx.New(errorx.BadRequest, "Invalid body")
	}

	if err := parseSession(ctx, req); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot find the session: %v", err)
		return errorx.New(errorx.Internal, "Cannot find the session")
	}

	return nil
}

func runMiddleware(
	ctx context.Context,
	befores, afters []MiddlewareFunc,
	closers []CloserFunc,
	handler func(context.Context) (any, error),
) {
	func() {
		if xcontext.Error(ctx) != nil {
			return
		}

		for _, m := range befores {
			rctx, err := m(ctx)
			if err != nil {
				ctx = xcontext.WithError(ctx, err)
				return
			}

			if rctx != nil {
				ctx = rctx
			}
		}

		resp, err := handler(ctx)
		if err != nil {
			ctx = xcontext.WithError(ctx, err)
			return
		}

		if resp != nil {
			ctx = xcontext.WithResponse(ctx, resp)
		}

		for _, m := range afters {
			rctx, err := m(ctx)
			if err != nil {
				ctx = xcontext.WithError(ctx, err)
				return
			}

			if rctx != nil {
				ctx = rctx
			}
		}
	}()

	for _, closer := range closers {
		closer(ctx)
	}
}
