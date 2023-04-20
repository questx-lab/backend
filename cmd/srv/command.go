package main

import "github.com/urfave/cli/v2"

// NewApp creates an app with sane defaults.
func (s *srv) loadApp() {
	s.app = cli.NewApp()
	s.app.Action = cli.ShowAppHelp
	s.app.Name = "Xquest"
	s.app.Usage = ""
	s.app.Commands = []*cli.Command{
		{
			Action:      server.startApi,
			Name:        "api",
			Usage:       "Start service api",
			ArgsUsage:   "<genesisPath>",
			Flags:       []cli.Flag{},
			Category:    "Api",
			Description: `Used for start service api, it main service included all apis.`,
		},
		{
			Action:      server.startGameProxy,
			Name:        "game_proxy",
			Usage:       "Start service game proxy",
			ArgsUsage:   "<genesisPath>",
			Flags:       []cli.Flag{},
			Category:    "Websocket",
			Description: `Used to direct connection to client via websocket.`,
		},
		{
			Action:      server.startGameEngine,
			Name:        "game_engine",
			Usage:       "Start service game engine",
			ArgsUsage:   "<genesisPath>",
			Flags:       []cli.Flag{},
			Category:    "Worker",
			Description: `Used to start service game engine.`,
		},
	}
}
