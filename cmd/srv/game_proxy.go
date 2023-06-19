package main

import (
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

func (s *srv) startGameProxy(*cli.Context) error {
	s.ctx = xcontext.WithDB(s.ctx, s.newDatabase())
	s.migrateDB()
	s.loadStorage()
	s.loadRepos()
	s.loadPublisher()
	s.loadGame()
	s.loadDomains()
	s.loadGameProxyRouter()

	cfg := xcontext.Configs(s.ctx)
	httpSrv := &http.Server{
		Addr:    cfg.GameProxyServer.Address(),
		Handler: s.router.Handler(cfg.GameProxyServer),
	}

	subscriber := kafka.NewSubscriber(
		"proxy/"+uuid.NewString(),
		[]string{cfg.Kafka.Addr},
		[]string{model.GameActionResponseTopic},
		s.proxyRouter.Subscribe,
	)

	go subscriber.Subscribe(s.ctx)

	xcontext.Logger(s.ctx).Infof("Server start in port : %v", cfg.GameProxyServer.Port)
	if err := httpSrv.ListenAndServe(); err != nil {
		panic(err)
	}
	xcontext.Logger(s.ctx).Infof("Server stop")

	return nil
}

func (s *srv) loadGameProxyRouter() {
	cfg := xcontext.Configs(s.ctx)
	s.router = router.New(s.ctx)
	s.router.AddCloser(middleware.Logger(cfg.Env))
	s.router.Before(middleware.NewAuthVerifier().WithAccessToken().Middleware())
	router.Websocket(s.router, "/game", s.gameProxyDomain.ServeGameClient)
}

func (s *srv) loadGame() {
	s.proxyRouter = gameproxy.NewRouter()
}
