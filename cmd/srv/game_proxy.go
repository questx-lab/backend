package main

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/domain/gameproxy"
	"github.com/questx-lab/backend/internal/middleware"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/kafka"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/urfave/cli/v2"
)

func (app *App) startGameProxy(*cli.Context) error {
	app.loadStorage()
	app.loadRepos()
	app.loadPublisher()
	app.loadGame()
	app.loadDomains()
	app.loadGameProxyRouter()

	cfg := xcontext.Configs(app.ctx)
	app.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.GameProxyServer.Port),
		Handler: app.router.Handler(cfg.GameProxyServer),
	}

	responseSubscriber := kafka.NewSubscriber(
		"proxy/"+uuid.NewString(),
		[]string{cfg.Kafka.Addr},
		[]string{string(model.ResponseTopic)},
		app.proxyRouter.Subscribe,
	)

	go responseSubscriber.Subscribe(app.ctx)

	xcontext.Logger(app.ctx).Infof("Server start in port : %v", cfg.GameProxyServer.Port)
	if err := app.server.ListenAndServe(); err != nil {
		panic(err)
	}
	xcontext.Logger(app.ctx).Infof("Server stop")

	return nil
}

func (app *App) loadGameProxyRouter() {
	app.router = router.New(app.ctx)
	app.router.AddCloser(middleware.Logger())
	app.router.Before(middleware.NewAuthVerifier().WithAccessToken().Middleware())
	router.Websocket(app.router, "/game", app.gameProxyDomain.ServeGameClient)
}

func (app *App) loadGame() {
	app.proxyRouter = gameproxy.NewRouter()
}
