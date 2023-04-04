package middleware

import (
	"net/http"

	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type CookieResponse interface {
	CookieInfo(xcontext.Context) []http.Cookie
}

func HandleSetAccessToken() router.MiddlewareFunc {
	return func(ctx xcontext.Context) error {
		tokenResp, ok := xcontext.GetResponse(ctx).(CookieResponse)
		if ok {
			for _, cookie := range tokenResp.CookieInfo(ctx) {
				http.SetCookie(ctx.Writer(), &cookie)
			}
		}

		return nil
	}
}
