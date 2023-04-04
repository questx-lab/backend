package main

var server srv

func main() {
	server.loadConfig()
	// server.loadDatabase()
	server.loadRepos()
	server.loadDomains()
	server.loadRouter()
	server.startServer()
}
