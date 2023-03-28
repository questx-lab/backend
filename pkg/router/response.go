package router

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/questx-lab/backend/pkg/errorx"
)

type response struct {
	Errno int64  `json:"errno"`
	Error string `json:"error,omitempty"`
	Data  any    `json:"data,omitempty"`
}

func newResponse(data any) response {
	return response{
		Errno: 0,
		Data:  data,
	}
}

func newErrorResponse(err error) response {
	errx := errorx.Error{}
	if errors.As(err, &errx) {
		return response{
			Errno: int64(errx.Code),
			Error: errx.Message,
		}
	}

	return response{
		Errno: int64(errorx.Unknown.Code),
		Error: errorx.Unknown.Message,
	}
}

func handleResponse() CloserFunc {
	return func(ctx Context) {
		err := func() error {
			if err := ctx.Error(); err != nil {
				return err
			}

			if resp := ctx.GetResponse(); resp != nil {
				if err := writeJson(ctx.Writer(), newResponse(resp)); err != nil {
					ctx.Logger().Errorf("cannot write the response %v", err)
					return errorx.New(errorx.BadResponse, "Cannot write the response")
				}
			}

			return nil
		}()

		if err != nil {
			resp := newErrorResponse(err)
			if err := writeJson(ctx.Writer(), resp); err != nil {
				ctx.Logger().Errorf("cannot write the response: %s", err.Error())
			}
		}
	}
}

func writeJson(r http.ResponseWriter, resp any) error {
	b, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	if _, err := r.Write(b); err != nil {
		return err
	}

	return nil
}
