package srv

import (
	"context"

	"github.com/questx-lab/backend/idl/pb"
	"github.com/questx-lab/backend/internal/deliveries"
	"github.com/questx-lab/backend/internal/domains"
	"github.com/questx-lab/backend/internal/repositories"
	"github.com/questx-lab/backend/pkg/grpc_client"
)

func (s *server) loadLogger() error {
	// s.logger = logger.NewZapLogger("INFO", true)
	return nil
}

func (s *server) loadConfig(ctx context.Context) error {
	return nil
}

func (s *server) loadClients(ctx context.Context) error {
	//* load grpc clients

	defaultOptions := &grpc_client.Options{
		IsEnableRetry:     true,
		IsEnableValidator: true,
	}

	s.userConnClient = &grpc_client.ConnClient{
		ServiceName: "User",
		Options:     defaultOptions,
	}
	s.userClient = pb.NewUserServiceClient(s.userConnClient.Conn)
	s.authClient = pb.NewAuthServiceClient(s.userConnClient.Conn)

	s.factories = append(s.factories, s.userConnClient)

	return nil
}

func (s *server) loadRepositories() error {
	s.userRepo = repositories.NewUserRepository()
	return nil
}

func (s *server) loadServices() error {
	s.authDomain = domains.NewAuthDomain(s.userRepo)
	return nil
}

func (s *server) loadDeliveries() error {
	s.authDelivery = deliveries.NewAuthDelivery(s.authDomain)
	return nil
}
