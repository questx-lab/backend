package domain

import (
	"errors"
	"log"

	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/api"
)

type UserDomain interface {
	GetUser(*api.Context, *model.GetUserRequest) (*model.GetUserResponse, error)
}

type userDomain struct {
	userRepo repository.UserRepository
}

func NewUserDomain(userRepo repository.UserRepository) UserDomain {
	return &userDomain{
		userRepo: userRepo,
	}
}

func (d *userDomain) GetUser(
	ctx *api.Context, req *model.GetUserRequest,
) (*model.GetUserResponse, error) {
	id := ctx.GetUserID()
	if id == "" || id != req.ID {
		return nil, errors.New("permission denied")
	}

	user, err := d.userRepo.RetrieveByID(ctx, req.ID)
	if err != nil {
		log.Println("Cannot get the user, err = ", err)
		return nil, err
	}

	return &model.GetUserResponse{
		ID:      user.ID,
		Address: user.Address,
		Name:    user.Name,
	}, nil
}
