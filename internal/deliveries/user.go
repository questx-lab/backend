package deliveries

import (
	"context"

	"github.com/questx-lab/backend/idl/pb"
	"github.com/questx-lab/backend/internal/domains"
)

type UserDelivery struct {
	userDomain domains.UserDomain
	pb.UnimplementedUserServiceServer
}

func NewUserDelivery(userDomain domains.UserDomain) pb.UserServiceServer {
	return &UserDelivery{
		userDomain: userDomain,
	}
}

func (d *UserDelivery) Login(_ context.Context, _ *pb.LoginRequest) (*pb.LoginResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (d *UserDelivery) Register(_ context.Context, _ *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	panic("not implemented") // TODO: Implement
}
