package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/questx-lab/backend/internal/domain/search"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/urfave/cli/v2"
)

func (s *srv) startSearchRPC(*cli.Context) error {
	indexer := search.NewBleveIndex(s.ctx)
	defer indexer.Close()
	go func() {
		termSignal := make(chan os.Signal, 1)
		signal.Notify(termSignal, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
		for sig := range termSignal {
			indexer.Close()
			xcontext.Logger(s.ctx).Errorf("Got a signal of %s", sig.String())
			os.Exit(1)
		}
	}()

	searchServerCfg := xcontext.Configs(s.ctx).SearchServer
	rpcServerName := searchServerCfg.RPCName
	rpcHandler := rpc.NewServer()
	err := rpcHandler.RegisterName(rpcServerName, indexer)
	if err != nil {
		xcontext.Logger(s.ctx).Infof("Cannot register indexer: %v", err)
		return err
	}
	defer rpcHandler.Stop()

	httpSrv := &http.Server{
		Handler: rpcHandler,
		Addr:    searchServerCfg.Address(),
	}

	xcontext.Logger(s.ctx).Infof("Started rpc server of search index")
	if err := httpSrv.ListenAndServe(); err != nil {
		xcontext.Logger(s.ctx).Errorf("A error occurs when running rpc server: %v", err)
		return err
	}
	xcontext.Logger(s.ctx).Infof("Stopped rpc server of search index")

	return nil
}
