package domain

import (
	"fmt"

	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/router"
)

type UserDomain interface {
	GetUser(router.Context, *model.GetUserRequest) (*model.GetUserResponse, error)
}

type userDomain struct {
	userRepo repository.UserRepository
}

func NewUserDomain(userRepo repository.UserRepository) UserDomain {
	return &userDomain{
		userRepo: userRepo,
	}
}

func (d *userDomain) GetUser(ctx router.Context, req *model.GetUserRequest) (*model.GetUserResponse, error) {
	user, err := d.userRepo.RetrieveByID(ctx, ctx.GetUserID())
	if err != nil {
		return nil, fmt.Errorf("%v: %w", err, errorx.ErrGeneric)
	}

	return &model.GetUserResponse{
		ID:      user.ID,
		Address: user.Address,
		Name:    user.Name,
	}, nil
}
