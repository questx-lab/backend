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
	collaboratorRepo repository.CollaboratorRepository
	userRepo         repository.UserRepository
}

func NewCommunityRoleVerifier(
	collaboratorRepo repository.CollaboratorRepository,
	userRepo repository.UserRepository,
) *CommunityRoleVerifier {
	return &CommunityRoleVerifier{
		collaboratorRepo: collaboratorRepo,
		userRepo:         userRepo,
	}
}

func (verifier *CommunityRoleVerifier) Verify(
	ctx context.Context,
	communityID string,
	requiredRoles ...entity.CollaboratorRole,
) error {
	userID := xcontext.RequestUserID(ctx)
	u, err := verifier.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user is not valid")
	}

	if u.Role == entity.RoleSuperAdmin || u.Role == entity.RoleAdmin {
		return nil
	}

	collaborator, err := verifier.collaboratorRepo.Get(ctx, communityID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user does not have permission")
		}

		return err
	}

	if !slices.Contains(requiredRoles, collaborator.Role) {
		return errors.New("user role does not have permission")
	}

	return nil
}
