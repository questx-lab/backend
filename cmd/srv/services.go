package srv

import (
	"context"
	"sisu-network/gateway/idl/pb"
	"sisu-network/gateway/pkg/http_server"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
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
