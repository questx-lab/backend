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
			Name:        "gateway",
			Usage:       "Bootstrap and start worker server",
			ArgsUsage:   "<genesisPath>",
			Flags:       []cli.Flag{},
			Category:    "Crawler Worker",
			Description: `Used to start crawler worker, clone data from omada cloud`,
		},
		{
			Action:      server.startWsProxy,
			Name:        "proxy",
			Usage:       "Bootstrap and start worker server",
			ArgsUsage:   "<genesisPath>",
			Flags:       []cli.Flag{},
			Category:    "Crawler Worker",
			Description: `Used to start crawler worker, clone data from omada cloud`,
		},
		{
			Action:      server.startSubscriber,
			Name:        "subscriber",
			Usage:       "Bootstrap and start worker server",
			ArgsUsage:   "<genesisPath>",
			Flags:       []cli.Flag{},
			Category:    "Crawler Worker",
			Description: `Used to start crawler worker, clone data from omada cloud`,
		},
	}
}
