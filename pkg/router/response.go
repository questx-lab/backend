package router

import (
	"encoding/json"
	"errors"
	"fmt"
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
	if errors.As(err, &errx) || errors.As(errors.Unwrap(err), &errx) {
		return response{
			Errno: int64(errx.Code),
			Error: errx.Message,
		}
	}

	return response{
		Errno: -1,
		Error: "Unknown Error",
	}
}

func handleResponse() CloserFunc {
	return func(ctx Context) {
		err := func() error {
			if err := ctx.Error(); err != nil {
				return err
			}

			resp := ctx.GetResponse()
			if resp == nil {
				return fmt.Errorf("no response: %w", errorx.ErrBadResponse)
			}

			if err := writeJson(ctx.Writer(), newResponse(resp)); err != nil {
				return fmt.Errorf("%v: %w", err, errorx.ErrBadResponse)
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
