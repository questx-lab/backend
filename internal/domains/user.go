package domains

import (
	"sisu-network/gateway/internal/repositories"
)

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
