package domain

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/authenticator"
	"github.com/questx-lab/backend/pkg/xcontext"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

type projectRoleVerifier struct {
	collaboratorRepo repository.CollaboratorRepository
}

func newProjectRoleVerifier(collaboratorRepo repository.CollaboratorRepository) *projectRoleVerifier {
	return &projectRoleVerifier{collaboratorRepo: collaboratorRepo}
}

func (verifier *projectRoleVerifier) Verify(
	ctx xcontext.Context,
	projectID string,
	requiredRole ...entity.Role,
) error {
	userID := xcontext.GetRequestUserID(ctx)
	collaborator, err := verifier.collaboratorRepo.Get(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user does not have permission")
		}

		return err
	}

	if !slices.Contains(requiredRole, collaborator.Role) {
		return errors.New("user role does not have permission")
	}

	return nil
}

func generateRandomString() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

func generateUniqueServiceUserID(authCfg authenticator.IOAuth2Config, serviceUserID string) string {
	return fmt.Sprintf("%s_%s", authCfg.Service(), serviceUserID)
}
