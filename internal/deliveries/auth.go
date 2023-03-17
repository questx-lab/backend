package deliveries

import (
	"context"
	"sisu-network/gateway/idl/pb"
	"sisu-network/gateway/internal/domains"
)

type AuthDelivery struct {
	authDomain domains.AuthDomain
	pb.UnimplementedAuthServiceServer
}

func NewAuthDelivery(authDomain domains.AuthDomain) pb.AuthServiceServer {
	return &AuthDelivery{
		authDomain: authDomain,
	}
}

func (d *AuthDelivery) Login(_ context.Context, _ *pb.LoginRequest) (*pb.LoginResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (d *AuthDelivery) Register(_ context.Context, _ *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	panic("not implemented") // TODO: Implement
}
