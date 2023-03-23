package router

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gorilla/sessions"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/authenticator"
)

type HandlerFunc[Request, Response any] func(ctx Context, req Request) (*Response, error)
type MiddlewareFunc func(ctx Context) error

type Router struct {
	mux *http.ServeMux

	before []MiddlewareFunc
	after  []MiddlewareFunc

	cfg               config.Configs
	accessTokenEngine authenticator.TokenEngine[model.AccessToken]
	sessionStore      sessions.Store
}

func New(cfg config.Configs) *Router {
	r := &Router{
		mux:               http.NewServeMux(),
		cfg:               cfg,
		accessTokenEngine: authenticator.NewTokenEngine[model.AccessToken](cfg.Token),
		sessionStore:      sessions.NewCookieStore([]byte(cfg.Session.Secret)),
	}

	r.After(handleResponse())
	return r
}

func GET[Request, Response any](router *Router, pattern string, handler HandlerFunc[Request, Response]) {
	route(router, http.MethodGet, pattern, handler)
}

func POST[Request, Response any](router *Router, pattern string, handler HandlerFunc[Request, Response]) {
	route(router, http.MethodPost, pattern, handler)
}

func route[Request, Response any](router *Router, method, pattern string, handler HandlerFunc[Request, Response]) {
	before := make([]MiddlewareFunc, len(router.before))
	after := make([]MiddlewareFunc, len(router.after))

	copy(before, router.before)
	copy(after, router.after)

	router.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if method != r.Method {
			http.Error(w, "unsupported method "+r.Method, http.StatusBadRequest)
			return
		}

		var req Request
		err := parseBody(r, &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx := &defaultContext{
			Context:           r.Context(),
			r:                 r,
			w:                 w,
			configs:           router.cfg,
			accessTokenEngine: router.accessTokenEngine,
			sessionStore:      router.sessionStore,
		}

		err = parseSession(ctx, &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for i := range before {
			if err := before[i](ctx); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		resp, err := handler(ctx, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if resp != nil {
			ctx.SetResponse(resp)
		}

		for i := range after {
			if err := after[i](ctx); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	})
}

func (r *Router) Before(middleware MiddlewareFunc) {
	r.before = append(r.before, middleware)
}

func (r *Router) After(middleware MiddlewareFunc) {
	r.after = append([]MiddlewareFunc{middleware}, r.after...)
}

func (r *Router) Branch() *Router {
	clone := &Router{
		mux:               r.mux,
		cfg:               r.cfg,
		accessTokenEngine: r.accessTokenEngine,
		sessionStore:      r.sessionStore,
		before:            make([]MiddlewareFunc, len(r.before)),
		after:             make([]MiddlewareFunc, len(r.after)),
	}
	copy(clone.before, r.before)
	copy(clone.after, r.after)

	return clone
}

func (r *Router) Static(root, relativePath string) {
	r.mux.Handle(root, http.FileServer(http.Dir(relativePath)))
}

func (r *Router) Handler() http.Handler {
	return r.mux
}

func parseBody(r *http.Request, req any) error {
	switch r.Method {
	case http.MethodGet:
		v := reflect.ValueOf(req).Elem()
		for i := 0; i < v.NumField(); i++ {
			name := v.Type().Field(i).Tag.Get("json")
			queryVal := r.URL.Query().Get(name)
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
			}
		}

	case http.MethodPost:
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(b, &req); err != nil {
			return err
		}

	default:
		return errors.New("unsupported method")
	}

	return nil
}

func parseSession(ctx Context, req any) error {
	session, err := ctx.SessionStore().Get(ctx.Request(), ctx.Configs().Session.Name)
	if err != nil {
		return err
	}

	v := reflect.ValueOf(req).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		name := v.Type().Field(i).Tag.Get("session")
		if name == "" {
			continue
		}

		value, ok := session.Values[name]
		if !ok {
			return errors.New("session value not found")
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

func handleResponse() MiddlewareFunc {
	return func(ctx Context) error {
		resp := ctx.GetResponse()
		if resp == nil {
			return nil
		}

		b, err := json.Marshal(resp)
		if err != nil {
			return err
		}

		if _, err := ctx.Writer().Write(b); err != nil {
			return err
		}
		return nil
	}
}
