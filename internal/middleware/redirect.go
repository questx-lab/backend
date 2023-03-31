package middleware

import (
	"net/http"

	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type RedirectResponse interface {
	RedirectInfo() (int, string)
}

func HandleRedirect() router.MiddlewareFunc {
	return func(ctx xcontext.Context) error {
		redirectResp, ok := ctx.GetResponse().(RedirectResponse)
		if !ok {
			return nil
		}

		code, uri := redirectResp.RedirectInfo()
		http.Redirect(ctx.Writer(), ctx.Request(), uri, code)

		// After rendering redirect response, do not render another response to client.
		ctx.SetResponse(nil)

		return nil
	}
}
