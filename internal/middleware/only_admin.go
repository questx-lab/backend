package middleware

import (
	"context"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/router"
)

type OnlyAdmin struct {
	globalRoleVerifier *common.GlobalRoleVerifier
}

func NewOnlyAdmin(userRepo repository.UserRepository) *OnlyAdmin {
	return &OnlyAdmin{
		globalRoleVerifier: common.NewGlobalRoleVerifier(userRepo),
	}
}

func (a *OnlyAdmin) Middleware() router.MiddlewareFunc {
	return func(ctx context.Context) (context.Context, error) {
		if err := a.globalRoleVerifier.Verify(ctx, entity.GlobalAdminRoles...); err != nil {
			return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
		}

		return nil, nil
	}
}
