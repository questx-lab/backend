package main

import (
	"os"
)

var server srv

func main() {
	server.loadConfig()
	server.loadLogger()
	server.loadEndpoint()
	// server.loadDatabase()
	server.loadAuthVerifier()
	// server.loadStorage()
	server.loadRepos()
	server.loadPublisher()
	server.loadDomains()
	server.loadSubscriber()
	server.loadApp()
	if err := server.app.Run(os.Args); err != nil {
		panic(err)
	}
}
