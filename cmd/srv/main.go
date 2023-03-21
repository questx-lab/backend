package main

var server srv

func main() {
	server.loadConfig()
	server.loadMux()
	server.loadDatabase()
	server.loadRepos()
	server.loadDomains()
	server.loadControllers()
	server.startServer()
}
