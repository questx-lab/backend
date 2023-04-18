package main

import "github.com/urfave/cli/v2"

// NewApp creates an app with sane defaults.
func (s *srv) loadApp() {
	app := cli.NewApp()
	app.Action = cli.ShowAppHelp
	app.Name = "Xquest"
	app.Usage = ""
	app.Commands = []*cli.Command{
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
			Action:      server.startWsProxy,
			Name:        "proxy",
			Usage:       "Start service proxy",
			ArgsUsage:   "<genesisPath>",
			Flags:       []cli.Flag{},
			Category:    "Websocket",
			Description: `Used to direct connection to client via websocket.`,
		},
		{
			Action:      server.startWsSubscriber,
			Name:        "subscriber",
			Usage:       "Start service subscriber",
			ArgsUsage:   "<genesisPath>",
			Flags:       []cli.Flag{},
			Category:    "Worker",
			Description: `Used to start worker that executes commands to subscribe to the message queue.`,
		},
	}
}
