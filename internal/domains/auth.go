package domains

import (
	"context"

	"sisu-network/gateway/idl/pb"
	"sisu-network/gateway/internal/repositories"
)

type AuthDomain interface {
	Login(ctx context.Context, data *pb.LoginRequest) (*pb.LoginResponse, error)
	Register(ctx context.Context, data *pb.RegisterRequest) (*pb.RegisterResponse, error)
}

type authDomain struct {
	userRepo repositories.UserRepository
}

func NewAuthDomain(userRepo repositories.UserRepository) AuthDomain {
	return &authDomain{
		userRepo: userRepo,
	}
}

func (d *authDomain) Login(ctx context.Context, data *pb.LoginRequest) (*pb.LoginResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (d *authDomain) Register(ctx context.Context, data *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	panic("not implemented") // TODO: Implement
}
