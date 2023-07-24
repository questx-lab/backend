package common

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/xcontext"
	"golang.org/x/exp/slices"

	mathUtil "github.com/pkg/math"
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
	effectRoleIDs ...string,
) error {
	userID := xcontext.RequestUserID(ctx)

	for _, roleID := range effectRoleIDs {
		if slices.Contains([]string{entity.OwnerBaseRole, entity.UserBaseRole}, roleID) {
			return fmt.Errorf("unable to effect base role")
		}
	}
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

	ids := make([]string, 0, len(followerRoles))
	for _, followerRole := range followerRoles {
		ids = append(ids, followerRole.RoleID)
	}

	roles, err := verifier.roleRepo.GetByIDs(ctx, append(ids, effectRoleIDs...))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Unable to get roles by ids: %v", err)
		return err
	}

	var totalPermission uint64
	minPriority := math.MaxInt
	minEffectPriority := math.MaxInt

	for _, role := range roles {
		if slices.Contains(ids, role.ID) {
			totalPermission |= role.Permissions
			minPriority = mathUtil.MinInt(minPriority, role.Priority)
		} else {
			minEffectPriority = mathUtil.MinInt(minEffectPriority, role.Priority)
		}
	}
	if minPriority > minEffectPriority {
		return fmt.Errorf("unable to effect more priority role")
	}
	path := xcontext.HTTPRequest(ctx).URL.Path
	permission := entity.RBAC[path]

	if totalPermission&uint64(permission) == 0 {
		return fmt.Errorf("user does not have permission")
	}

	return nil
}
