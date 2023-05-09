package common

import (
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

func (verifier *GlobalRoleVerifier) Verify(ctx xcontext.Context, requiredRoles ...entity.GlobalRole) error {
	userID := xcontext.GetRequestUserID(ctx)
	u, err := verifier.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user is not valid")
	}

	if !slices.Contains(requiredRoles, u.Role) {
		return errors.New("user role does not have permission")
	}

	return nil
}

type ProjectRoleVerifier struct {
	collaboratorRepo repository.CollaboratorRepository
	userRepo         repository.UserRepository
}

func NewProjectRoleVerifier(
	collaboratorRepo repository.CollaboratorRepository,
	userRepo repository.UserRepository,
) *ProjectRoleVerifier {
	return &ProjectRoleVerifier{
		collaboratorRepo: collaboratorRepo,
		userRepo:         userRepo,
	}
}

func (verifier *ProjectRoleVerifier) Verify(
	ctx xcontext.Context,
	projectID string,
	requiredRoles ...entity.Role,
) error {
	userID := xcontext.GetRequestUserID(ctx)
	u, err := verifier.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user is not valid")
	}

	if u.Role == entity.RoleSuperAdmin || u.Role == entity.RoleAdmin {
		return nil
	}

	collaborator, err := verifier.collaboratorRepo.Get(ctx, projectID, userID)
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
