package common

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"math/big"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/xcontext"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

type ProjectRoleVerifier struct {
	collaboratorRepo repository.CollaboratorRepository
}

func NewProjectRoleVerifier(collaboratorRepo repository.CollaboratorRepository) *ProjectRoleVerifier {
	return &ProjectRoleVerifier{collaboratorRepo: collaboratorRepo}
}

func (verifier *ProjectRoleVerifier) Verify(
	ctx xcontext.Context,
	projectID string,
	requiredRoles ...entity.Role,
) error {
	userID := xcontext.GetRequestUserID(ctx)
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

func GenerateRandomString() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GenerateRandomAlphabet(n uint) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = alphabet[CryptoRandIntn(len(alphabet))]
	}
	return string(b)
}

func Hash(b []byte) string {
	hashed := sha256.Sum224(b)
	return base64.StdEncoding.EncodeToString(hashed[:])
}

// CryptoRandIntn returns a uniform random value in [0, n). It panics if got a negative parameter.
func CryptoRandIntn(n int) int {
	r, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		panic(err)
	}

	return int(r.Int64())
}
