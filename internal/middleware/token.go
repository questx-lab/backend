package middleware

import (
	"net/http"
	"time"

	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type AccessTokenResponse interface {
	AccessTokenInfo() string
}

func HandleSetAccessToken() router.MiddlewareFunc {
	return func(ctx xcontext.Context) error {
		tokenResp, ok := ctx.GetResponse().(AccessTokenResponse)
		if ok {
			accessToken := tokenResp.AccessTokenInfo()
			http.SetCookie(ctx.Writer(), &http.Cookie{
				Name:     ctx.Configs().Auth.AccessTokenName,
				Value:    accessToken,
				Domain:   "",
				Path:     "/",
				Expires:  time.Now().Add(ctx.Configs().Token.Expiration),
				Secure:   true,
				HttpOnly: false,
			})
		}

		return nil
	}
}
