package middleware

import (
	"fmt"

	"github.com/questx-lab/backend/pkg/router"
)

func Logger() router.CloserFunc {
	return func(ctx router.Context) {
		info := fmt.Sprintf("%s | %s", ctx.Request().Method, ctx.Request().URL.Path)
		if err := ctx.Error(); err != nil {
			ctx.Logger().Errorf("%s\n%s", info, err.Error())
		} else {
			ctx.Logger().Infof(info)
		}
	}
}
