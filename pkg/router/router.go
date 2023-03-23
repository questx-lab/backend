package router

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/authenticator"
	"github.com/questx-lab/backend/pkg/errorx"
)

type HandlerFunc[Request, Response any] func(ctx Context, req *Request) (*Response, error)
type MiddlewareFunc func(ctx Context) error
type CloserFunc func(ctx Context)

type Router struct {
	mux *http.ServeMux

	befores []MiddlewareFunc
	afters  []MiddlewareFunc
	closers []CloserFunc

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

	r.AddCloser(handleResponse())
	return r
}

func GET[Request, Response any](router *Router, pattern string, handler HandlerFunc[Request, Response]) {
	route(router, http.MethodGet, pattern, handler)
}

func POST[Request, Response any](router *Router, pattern string, handler HandlerFunc[Request, Response]) {
	route(router, http.MethodPost, pattern, handler)
}

func route[Request, Response any](router *Router, method, pattern string, handler HandlerFunc[Request, Response]) {
	befores := make([]MiddlewareFunc, len(router.befores))
	afters := make([]MiddlewareFunc, len(router.afters))
	closers := make([]CloserFunc, len(router.closers))

	copy(befores, router.befores)
	copy(afters, router.afters)
	copy(closers, router.closers)

	router.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r.Context(), r, w, router.cfg)

		var req Request
		err := func() error {
			if method != r.Method {
				return fmt.Errorf("%s: %w", r.Method, errorx.ErrNotSupportedMethod)
			}

			err := parseBody(r, &req)
			if err != nil {
				return fmt.Errorf("%v: %w", err, errorx.ErrBadRequest)
			}

			err = parseSession(ctx, &req)
			if err != nil {
				return fmt.Errorf("%v: %w", err, errorx.ErrBadRequest)
			}

			return nil
		}()

		func() {
			if err != nil {
				ctx.SetError(err)
				return
			}

			for _, m := range befores {
				if err := m(ctx); err != nil {
					ctx.SetError(err)
					return
				}
			}

			if ctx.Error() == nil {
				resp, err := handler(ctx, &req)
				if err != nil {
					ctx.SetError(err)
					return
				} else if resp != nil {
					ctx.SetResponse(resp)
				}
			}

			for _, m := range afters {
				if err := m(ctx); err != nil {
					ctx.SetError(err)
					return
				}
			}
		}()

		for _, c := range closers {
			c(ctx)
		}
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

func (r *Router) Branch() *Router {
	clone := &Router{
		mux:               r.mux,
		cfg:               r.cfg,
		accessTokenEngine: r.accessTokenEngine,
		sessionStore:      r.sessionStore,
		befores:           make([]MiddlewareFunc, len(r.befores)),
		afters:            make([]MiddlewareFunc, len(r.afters)),
		closers:           make([]CloserFunc, len(r.closers)),
	}
	copy(clone.befores, r.befores)
	copy(clone.afters, r.afters)
	copy(clone.closers, r.closers)

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
