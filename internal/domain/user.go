package domain

import "github.com/questx-lab/backend/internal/repository"

type UserDomain interface {
}

type userDomain struct {
	userRepo repository.UserRepository
}

func NewUserDomain(userRepo repository.UserRepository) UserDomain {
	return &userDomain{
		userRepo: userRepo,
	}
}
