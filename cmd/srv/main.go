package main

var server srv

func main() {
	server.loadMux()
	server.loadConfig()
	server.loadDatabase()
	server.loadRepos()
	server.loadDomains()
	server.loadControllers()
	server.startServer()
}
