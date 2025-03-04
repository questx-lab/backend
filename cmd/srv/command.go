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
		{
			Action:      s.startNotificationProxy,
			Name:        "notification_proxy",
			Usage:       "Start service notification proxy",
			ArgsUsage:   "<genesisPath>",
			Flags:       []cli.Flag{},
			Category:    "Notification",
			Description: `Used to direct connection to client via websocket.`,
		},
		{
			Action:      s.startNotificationEngine,
			Name:        "notification_engine",
			Usage:       "Start service notification engine",
			ArgsUsage:   "<genesisPath>",
			Flags:       []cli.Flag{},
			Category:    "Notification",
			Description: `Used to broadcast event.`,
		},
	}
	if err := cliapp.Run(os.Args); err != nil {
		panic(err)
	}
}
