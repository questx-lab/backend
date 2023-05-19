package middleware

import (
	"context"

	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/xcontext"
	"golang.org/x/exp/slices"
)

func MustUpdateUsername(userRepo repository.UserRepository, excludes ...string) router.MiddlewareFunc {
	return func(ctx context.Context) (context.Context, error) {
		requestUserID := xcontext.RequestUserID(ctx)
		if requestUserID != "" {
			if !slices.Contains(excludes, xcontext.HTTPRequest(ctx).URL.Path) {
				user, err := userRepo.GetByID(ctx, requestUserID)
				if err != nil {
					xcontext.Logger(ctx).Errorf("Cannot get user: %v", err)
					return nil, errorx.Unknown
				}

				if user.IsNewUser {
					return nil, errorx.New(errorx.MustUpdateUsername,
						"User must setup username before using application")
				}
			}
		}

		return nil, nil
	}
}
