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
		signal.Notify(termSignal, syscall.SIGINT, syscall.SIGTERM)
		for sig := range termSignal {
			indexer.Close()
			xcontext.Logger(s.ctx).Errorf("Got a signal of %s", sig.String())
			os.Exit(1)
		}
	}()

	searchServerCfg := xcontext.Configs(s.ctx).SearchServer
	rpcHandler := rpc.NewServer()
	defer rpcHandler.Stop()
	err := rpcHandler.RegisterName(searchServerCfg.RPCName, indexer)
	if err != nil {
		xcontext.Logger(s.ctx).Infof("Cannot register indexer: %v", err)
		return err
	}

	xcontext.Logger(s.ctx).Infof("Started rpc server of search index")
	httpSrv := &http.Server{
		Handler: rpcHandler,
		Addr:    searchServerCfg.Address(),
	}
	if err := httpSrv.ListenAndServe(); err != nil {
		xcontext.Logger(s.ctx).Errorf("An error occurs when running rpc server: %v", err)
		return err
	}

	return nil
}
