package domains

import "github.com/questx-lab/backend/internal/repositories"

type UserDomain interface {
}

type userDomain struct {
	userRepo repositories.UserRepository
}

func NewUserDomain(userRepo repositories.UserRepository) UserDomain {
	return &userDomain{
		userRepo: userRepo,
	}
}
