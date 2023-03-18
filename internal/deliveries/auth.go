package deliveries

import (
	"context"

	"github.com/questx-lab/backend/idl/pb"
	"github.com/questx-lab/backend/internal/domains"
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
