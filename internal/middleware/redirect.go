package middleware

import (
	"net/http"

	"github.com/questx-lab/backend/pkg/router"
)

type RedirectResponse interface {
	RedirectInfo() (int, string)
}

func HandleRedirect() router.MiddlewareFunc {
	return func(ctx router.Context) error {
		redirectResp, ok := ctx.GetResponse().(RedirectResponse)
		if !ok {
			return nil
		}

		code, uri := redirectResp.RedirectInfo()
		http.Redirect(ctx.Writer(), ctx.Request(), uri, code)
		ctx.OverrideResponse()

		return nil
	}
}
