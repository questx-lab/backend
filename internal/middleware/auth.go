package middleware

import (
	"errors"

	"github.com/questx-lab/backend/pkg/router"
)

func Authenticate() router.MiddlewareFunc {
	return func(ctx *router.Context) error {
		if ctx.GetUserID() == "" {
			return errors.New("you need to authenticate before")
		}
		return nil
	}
}
