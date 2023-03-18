package srv

import (
	"context"

	"github.com/questx-lab/backend/idl/pb"
	"github.com/questx-lab/backend/pkg/grpc_server"
	"github.com/questx-lab/backend/pkg/http_server"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

func Gateway(ctx context.Context) error {
	srv.httpServer = &http_server.HttpServer{
		Address: srv.configs.HTTP,
		Logger:  srv.logger,
		Handlers: func(ctx context.Context, mux *runtime.ServeMux) error {
			if err := pb.RegisterUserServiceHandlerClient(ctx, mux, srv.userClient); err != nil {
				return err
			}

			if err := pb.RegisterAuthServiceHandlerClient(ctx, mux, srv.authClient); err != nil {
				return err
			}

			return nil
		},
		Options: &http_server.Options{},
	}

	srv.processors = append(srv.processors, srv.httpServer)

	return nil
}

func User(ctx context.Context) error {
	srv.grpcServer = &grpc_server.GrpcServer{
		ServiceName: srv.configs.ServiceName,
		Address:     srv.configs.GRPC,
		Logger:      srv.logger,
		Handlers: func(ctx context.Context, server *grpc.Server) error {
			pb.RegisterUserServiceServer(server, srv.userDelivery)
			pb.RegisterAuthServiceServer(server, srv.authDelivery)
			return nil
		},
		Options: &grpc_server.Options{
			IsEnableRecovery:  true,
			IsEnableValidator: true,
		},
	}

	return nil
}
