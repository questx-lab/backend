package common

import (
	"context"
	"errors"
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/xcontext"
	"golang.org/x/exp/slices"
)

type GlobalRoleVerifier struct {
	userRepo repository.UserRepository
}

func NewGlobalRoleVerifier(userRepo repository.UserRepository) *GlobalRoleVerifier {
	return &GlobalRoleVerifier{userRepo: userRepo}
}

func (verifier *GlobalRoleVerifier) Verify(ctx context.Context, requiredRoles ...entity.GlobalRole) error {
	userID := xcontext.RequestUserID(ctx)
	u, err := verifier.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user is not valid")
	}

	if !slices.Contains(requiredRoles, u.Role) {
		return errors.New("user role does not have permission")
	}

	return nil
}

type CommunityRoleVerifier struct {
	followerRoleRepo repository.FollowerRoleRepository
	roleRepo         repository.RoleRepository
	userRepo         repository.UserRepository
}

func NewCommunityRoleVerifier(
	followerRoleRepo repository.FollowerRoleRepository,
	roleRepo repository.RoleRepository,
	userRepo repository.UserRepository,
) *CommunityRoleVerifier {
	return &CommunityRoleVerifier{
		followerRoleRepo: followerRoleRepo,
		roleRepo:         roleRepo,
		userRepo:         userRepo,
	}
}

func (verifier *CommunityRoleVerifier) Verify(
	ctx context.Context,
	communityID string,
) error {
	userID := xcontext.RequestUserID(ctx)
	u, err := verifier.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user is not valid")
	}

	if u.Role == entity.RoleSuperAdmin || u.Role == entity.RoleAdmin {
		return nil
	}

	followerRoles, err := verifier.followerRoleRepo.Get(ctx, userID, communityID)
	if err != nil {
		return err
	}

	var totalPermission uint64

	for _, followerRole := range followerRoles {
		role, err := verifier.roleRepo.GetByID(ctx, followerRole.RoleID)
		if err != nil {
			return err
		}

		totalPermission = totalPermission | role.Permissions
	}

	path := xcontext.HTTPRequest(ctx).URL.Path
	permission := entity.RBAC[path]

	if totalPermission&uint64(permission) == 0 {
		return fmt.Errorf("user does not have permission")
	}

	return nil
}
