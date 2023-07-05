package main

import (
	"context"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/questx-lab/backend/pkg/logger"
	"github.com/questx-lab/backend/pkg/token"
	"github.com/questx-lab/backend/pkg/xcontext"
)

func main() {
	// Set the timezone to UTC globally.
	os.Setenv("TZ", "")

	server := srv{}
	server.ctx = context.Background()
	server.ctx = xcontext.WithConfigs(server.ctx, server.loadConfig())
	server.ctx = xcontext.WithHTTPClient(server.ctx, http.DefaultClient)
	server.ctx = xcontext.WithLogger(server.ctx, logger.NewLogger(logger.INFO))
	server.ctx = xcontext.WithTokenEngine(server.ctx,
		token.NewEngine(xcontext.Configs(server.ctx).Auth.TokenSecret))
	server.ctx = xcontext.WithSessionStore(server.ctx,
		sessions.NewCookieStore([]byte(xcontext.Configs(server.ctx).Session.Secret)))

	server.run()
}
