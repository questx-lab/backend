package domain

import (
	"github.com/questx-lab/backend/api"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
)

type AuthDomain interface {
	Login(ctx api.CustomContext, data *model.LoginRequest) (*model.LoginResponse, error)
	Register(ctx api.CustomContext, data *model.RegisterRequest) (*model.RegisterResponse, error)
}

type authDomain struct {
	userRepo repository.UserRepository
}

func NewAuthDomain(userRepo repository.UserRepository) AuthDomain {
	return &authDomain{
		userRepo: userRepo,
	}
}

func (d *authDomain) Login(ctx api.CustomContext, data *model.LoginRequest) (*model.LoginResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (d *authDomain) Register(ctx api.CustomContext, data *model.RegisterRequest) (*model.RegisterResponse, error) {
	panic("not implemented") // TODO: Implement
}
