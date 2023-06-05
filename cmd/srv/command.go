package main

import (
	"os"

	"github.com/urfave/cli/v2"
)

func (s *srv) run() {
	cliapp := cli.NewApp()
	cliapp.Action = cli.ShowAppHelp
	cliapp.Name = "Xquest"
	cliapp.Usage = ""
	cliapp.Commands = []*cli.Command{
		{
			Action:      s.startApi,
			Name:        "api",
			Usage:       "Start service api",
			ArgsUsage:   "<genesisPath>",
			Flags:       []cli.Flag{},
			Category:    "Api",
			Description: `Used for start service api, it main service included all apis.`,
		},
		{
			Action:      s.startCron,
			Name:        "cron",
			Usage:       "Start cron jobs",
			ArgsUsage:   "<genesisPath>",
			Flags:       []cli.Flag{},
			Category:    "Cron",
			Description: `Used to start cron jobs.`,
		},
		{
			Action:      s.startGameProxy,
			Name:        "game_proxy",
			Usage:       "Start service game proxy",
			ArgsUsage:   "<genesisPath>",
			Flags:       []cli.Flag{},
			Category:    "Game",
			Description: `Used to direct connection to client via websocket.`,
		},
		{
			Action:      s.startGameEngine,
			Name:        "game_engine",
			Usage:       "Start service game engine",
			ArgsUsage:   "<genesisPath>",
			Flags:       []cli.Flag{},
			Category:    "Game",
			Description: `Used to start service game engine.`,
		},
		{
			Action:      s.startSearchRPC,
			Name:        "search",
			Usage:       "Start search rpc server",
			ArgsUsage:   "<genesisPath>",
			Flags:       []cli.Flag{},
			Category:    "Search",
			Description: `Used to start search rpc server.`,
		},
		{
			Action:      s.startBlockchain,
			Name:        "blockchain",
			Usage:       "Start blockchain server",
			ArgsUsage:   "<genesisPath>",
			Flags:       []cli.Flag{},
			Category:    "Search",
			Description: `Used to start blockchain server.`,
		},
	}

	if err := cliapp.Run(os.Args); err != nil {
		panic(err)
	}
}
