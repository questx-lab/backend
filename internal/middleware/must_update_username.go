package middleware

import (
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/xcontext"
	"golang.org/x/exp/slices"
)

func MustUpdateUsername(userRepo repository.UserRepository, excludes ...string) router.MiddlewareFunc {
	return func(ctx xcontext.Context) error {
		requestUserID := xcontext.GetRequestUserID(ctx)
		if requestUserID != "" {
			if !slices.Contains(excludes, ctx.Request().URL.Path) {
				user, err := userRepo.GetByID(ctx, requestUserID)
				if err != nil {
					ctx.Logger().Errorf("Cannot get user: %v", err)
					return errorx.Unknown
				}

				if user.IsNewUser {
					return errorx.New(errorx.Unavailable,
						"User must setup username before using application")
				}
			}

			xcontext.SetRequestUserID(ctx, requestUserID)
		}

		return nil
	}
}
