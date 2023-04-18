package main

import (
	"os"
)

var server srv

func main() {
	load()
	if err := server.app.Run(os.Args); err != nil {
		panic(err)
	}
}

func load() {
	server.loadConfig()
	server.loadLogger()
	server.loadEndpoint()
	server.loadDatabase()
	server.loadStorage()
	server.loadRepos()
	server.loadDomains()
	server.loadRouter()
	server.loadApp()
}
