package domain

import (
	"crypto/rand"
	"encoding/base64"
	"errors"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

func verifyProjectPermission(
	ctx xcontext.Context,
	collaboratorRepo repository.CollaboratorRepository,
	projectID string,
	appendRole ...entity.Role,
) string {
	userID := ctx.GetUserID()

	collaborator, err := collaboratorRepo.Get(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "User does not have permission"
		}

		ctx.Logger().Errorf("Cannot get the collaborator: %v", err)
		return errorx.Unknown.Message
	}
	writerRole := []entity.Role{
		entity.Owner,
		entity.Editor,
	}
	if !slices.Contains(append(writerRole, appendRole...), collaborator.Role) {
		return "User role does not have permission"
	}

	return ""
}

func generateRandomString() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	state := base64.StdEncoding.EncodeToString(b)

	return state, nil
}
