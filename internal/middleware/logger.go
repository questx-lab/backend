package middleware

import (
	"errors"
	"fmt"

	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/xcontext"
)

func Logger() router.CloserFunc {
	return func(ctx xcontext.Context) {
		info := fmt.Sprintf("%s | %s", ctx.Request().Method, ctx.Request().URL.Path)
		if err := ctx.Error(); err != nil {
			var errx errorx.Error
			if errors.As(err, &errx) {
				ctx.Logger().Warnf("%s | %d", info, errx.Code)
			} else {
				ctx.Logger().Errorf("%s | %d", info, -1)
			}
		} else {
			ctx.Logger().Infof(info)
		}
	}
}
