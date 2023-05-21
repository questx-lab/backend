package main

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/questx-lab/backend/pkg/authenticator"
	"github.com/questx-lab/backend/pkg/logger"
	"github.com/questx-lab/backend/pkg/xcontext"
)

func main() {
	app := App{}
	app.ctx = context.Background()
	app.ctx = xcontext.WithConfigs(app.ctx, app.loadConfig())
	app.ctx = xcontext.WithHTTPClient(app.ctx, http.DefaultClient)
	app.ctx = xcontext.WithDB(app.ctx, app.newDatabase())
	app.ctx = xcontext.WithLogger(app.ctx, logger.NewLogger())
	app.ctx = xcontext.WithTokenEngine(app.ctx,
		authenticator.NewTokenEngine(xcontext.Configs(app.ctx).Auth.TokenSecret))
	app.ctx = xcontext.WithSessionStore(app.ctx,
		sessions.NewCookieStore([]byte(xcontext.Configs(app.ctx).Session.Secret)))

	app.migrateDB()
	app.run()
}
