package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/jwt"
	"github.com/questx-lab/backend/pkg/session"
)

type Endpoint[Request, Response any] struct {
	Method string
	Path   string
	Before []Handler //! middleware before handle
	Handle func(*Context, *Request) (*Response, error)
	After  []Handler //! middleware after handle
}

func (e *Endpoint[Request, Response]) Register(
	mux *http.ServeMux,
	accessTokenEngine *jwt.Engine[model.AccessToken],
	sessionStore *session.Store,
	cfg config.Configs,
) {
	mux.HandleFunc(e.Path, func(w http.ResponseWriter, r *http.Request) {
		ctx := &Context{
			Context: r.Context(),
			r:       r,
			w:       w,

			AccessTokenEngine: accessTokenEngine,
			SessionStore:      sessionStore,
			Configs:           cfg,
		}

		for _, h := range e.Before {
			h(ctx)
		}

		var req Request
		e.readJson(ctx, &req)

		resp, err := e.Handle(ctx, &req)
		if err != nil {
			http.Error(ctx.w, err.Error(), http.StatusInternalServerError)
		} else {
			e.writeJson(ctx, resp)
		}

		for _, h := range e.After {
			h(ctx)
		}
	})
}

func (e *Endpoint[Request, Response]) readJson(ctx *Context, req *Request) {
	//* marshal step
	switch e.Method {
	case http.MethodGet, http.MethodDelete:
		v := reflect.ValueOf(req).Elem()
		for i := 0; i < v.NumField(); i++ {
			name := v.Type().Field(i).Tag.Get("json")
			queryVal := ctx.r.URL.Query().Get(name)
			pointer := v.Field(i).Addr().Interface()

			switch v.Field(i).Kind() {
			case reflect.String:
				p := pointer.(*string)
				*p = queryVal

			case reflect.Int:
				p := pointer.(*int)
				val, err := strconv.Atoi(queryVal)
				if err != nil {
					http.Error(ctx.w, err.Error(), http.StatusBadRequest)
				}
				*p = val
			}
		}

	case http.MethodPost, http.MethodPut, http.MethodPatch:
		b, err := ioutil.ReadAll(ctx.r.Body)
		if err != nil {
			http.Error(ctx.w, err.Error(), http.StatusBadRequest)
		}

		if err := json.Unmarshal(b, &req); err != nil {
			http.Error(ctx.w, err.Error(), http.StatusBadRequest)
		}
	}
}

func (e *Endpoint[Request, Response]) writeJson(ctx *Context, resp *Response) {
	b, err := json.Marshal(resp)
	if err != nil {
		http.Error(ctx.w, err.Error(), http.StatusInternalServerError)
	}

	if _, err := ctx.w.Write(b); err != nil {
		http.Error(ctx.w, err.Error(), http.StatusInternalServerError)
	}
}
