package middleware

import (
	"context"
	"errors"
	"fmt"

	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/xcontext"
)

func Logger(env string) router.CloserFunc {
	return func(ctx context.Context) {
		req := xcontext.HTTPRequest(ctx)
		info := fmt.Sprintf("%s | %s", req.Method, req.URL.Path)
		if err := xcontext.Error(ctx); err != nil {
			var errx errorx.Error
			if errors.As(err, &errx) {
				xcontext.Logger(ctx).Warnf("%s | %d", info, errx.Code)
			} else {
				xcontext.Logger(ctx).Errorf("%s | %d", info, -1)
			}
		} else {
			if env == "local" {
				xcontext.Logger(ctx).Infof(info)
			}
		}
	}
}
