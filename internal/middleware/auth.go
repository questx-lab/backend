package middleware

import (
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/router"
)

func Authenticate() router.MiddlewareFunc {
	return func(ctx router.Context) error {
		if ctx.GetUserID() == "" {
			return errorx.NewGeneric(nil, "You need to authenticate before")
		}
		return nil
	}
}
