package middleware

import (
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/xcontext"
)

func Authenticate() router.MiddlewareFunc {
	return func(ctx xcontext.Context) error {
		if ctx.GetUserID() == "" {
			return errorx.New(errorx.Unauthenticated, "You need to authenticate before")
		}
		return nil
	}
}
