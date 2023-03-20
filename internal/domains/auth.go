package domains

import (
	"github.com/questx-lab/backend/api"
	"github.com/questx-lab/backend/internal/models"
	"github.com/questx-lab/backend/internal/repositories"
)

type AuthDomain interface {
	Login(ctx api.CustomContext, data *models.LoginRequest) (*models.LoginResponse, error)
	Register(ctx api.CustomContext, data *models.RegisterRequest) (*models.RegisterResponse, error)
}

type authDomain struct {
	userRepo repositories.UserRepository
}

func NewAuthDomain(userRepo repositories.UserRepository) AuthDomain {
	return &authDomain{
		userRepo: userRepo,
	}
}

func (d *authDomain) Login(ctx api.CustomContext, data *models.LoginRequest) (*models.LoginResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (d *authDomain) Register(ctx api.CustomContext, data *models.RegisterRequest) (*models.RegisterResponse, error) {
	panic("not implemented") // TODO: Implement
}
