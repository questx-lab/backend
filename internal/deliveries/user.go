package deliveries

import (
	"context"
	"sisu-network/gateway/idl/pb"
	"sisu-network/gateway/internal/domains"
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
