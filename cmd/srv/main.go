package main

import (
	"context"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/questx-lab/backend/pkg/authenticator"
	"github.com/questx-lab/backend/pkg/logger"
	"github.com/questx-lab/backend/pkg/xcontext"
)

var server srv

func main() {
	server.ctx = context.Background()
	server.ctx = xcontext.WithConfigs(server.ctx, server.loadConfig())
	server.ctx = xcontext.WithHTTPClient(server.ctx, http.DefaultClient)
	server.ctx = xcontext.WithDB(server.ctx, server.newDatabase())
	server.ctx = xcontext.WithLogger(server.ctx, logger.NewLogger())
	server.ctx = xcontext.WithTokenEngine(server.ctx,
		authenticator.NewTokenEngine(xcontext.Configs(server.ctx).Auth.TokenSecret))
	server.ctx = xcontext.WithSessionStore(server.ctx,
		sessions.NewCookieStore([]byte(xcontext.Configs(server.ctx).Session.Secret)))

	server.loadApp()
	if err := server.app.Run(os.Args); err != nil {
		panic(err)
	}
}
