package srv

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/questx-lab/backend/idl/pb"
	"github.com/questx-lab/backend/internal/domains"
	"github.com/questx-lab/backend/internal/repositories"
	"github.com/questx-lab/backend/pkg/configs"
	"github.com/questx-lab/backend/pkg/grpc_client"
	"github.com/questx-lab/backend/pkg/grpc_server"
	"github.com/questx-lab/backend/pkg/http_server"

	"go.uber.org/zap"
)

var srv server

type server struct {
	configs *configs.Config

	//* load client connections
	userConnClient *grpc_client.ConnClient

	//* load clients
	userClient pb.UserServiceClient
	authClient pb.AuthServiceClient

	//* load repositories
	userRepo repositories.UserRepository

	//* load services
	authDomain domains.AuthDomain

	//* load deliveries
	authDelivery pb.AuthServiceServer
	userDelivery pb.UserServiceServer

	//* load servers
	httpServer *http_server.HttpServer
	grpcServer *grpc_server.GrpcServer

	//* logger
	logger *zap.Logger

	processors []processor
	factories  []factory
}

type processor interface {
	Init(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type factory interface {
	Connect(ctx context.Context) error
	Stop(ctx context.Context) error
}

func load(ctx context.Context) error {

	if err := srv.loadConfig(ctx); err != nil {
		return err
	}

	if err := srv.loadLogger(); err != nil {
		return err
	}

	if err := srv.loadClients(ctx); err != nil {
		return err
	}
	if err := srv.loadRepositories(); err != nil {
		return err
	}

	if err := srv.loadServices(); err != nil {
		return err
	}

	if err := srv.loadDeliveries(); err != nil {
		return err
	}

	return nil
}
func start(ctx context.Context) error {
	errChan := make(chan error)

	for _, f := range srv.factories {
		if err := f.Connect(ctx); err != nil {
			return err
		}
	}

	for _, p := range srv.processors {
		go func(p processor) {
			if err := p.Start(ctx); err != nil {
				errChan <- err
			}
		}(p)
	}
	go func() {
		err := <-errChan
		srv.logger.Sugar().Errorf("start error: %w\n", err)
	}()
	return nil
}

func stop(ctx context.Context) error {
	for _, processor := range srv.processors {
		if err := processor.Stop(ctx); err != nil {
			return err
		}
	}

	for _, database := range srv.factories {
		if err := database.Stop(ctx); err != nil {
			return err
		}
	}
	return nil
}

func GracefulShutdown(ctx context.Context, fn func(context.Context) error) error {
	// TODO: with graceful shutdown
	timeWait := 15 * time.Second
	signChan := make(chan os.Signal, 1)

	if err := load(ctx); err != nil {
		return err
	}

	if err := fn(ctx); err != nil {
		return err
	}
	signal.Notify(signChan, os.Interrupt, syscall.SIGTERM)
	<-signChan
	srv.logger.Sugar().Infoln("Shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), timeWait)
	defer func() {
		srv.logger.Sugar().Infoln("Close another connection")
		cancel()
	}()
	if err := stop(ctx); err == context.DeadlineExceeded {
		return fmt.Errorf("Halted active connections")
	}
	close(signChan)
	srv.logger.Sugar().Infoln("server down Completed")
	return nil
}
