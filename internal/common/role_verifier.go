package common

import (
	"context"
	"errors"
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/xcontext"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
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
	followerRepo repository.FollowerRepository
	roleRepo     repository.RoleRepository
	userRepo     repository.UserRepository
}

func NewCommunityRoleVerifier(
	followerRepo repository.FollowerRepository,
	roleRepo repository.RoleRepository,
	userRepo repository.UserRepository,
) *CommunityRoleVerifier {
	return &CommunityRoleVerifier{
		followerRepo: followerRepo,
		roleRepo:     roleRepo,
		userRepo:     userRepo,
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

	follower, err := verifier.followerRepo.Get(ctx, userID, communityID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user does not have permission")
		}

		return err
	}
	role, err := verifier.roleRepo.GetRoleByID(ctx, follower.RoleID)
	if err != nil {
		return err
	}

	path := xcontext.HTTPRequest(ctx).URL.Path
	permission := entity.RBAC[path]

	if role.Permissions&uint64(permission) == 0 {
		return fmt.Errorf("user does not have permission")
	}

	return nil
}
