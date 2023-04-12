package main

var server srv

func main() {
	server.loadConfig()
	server.loadEndpoint()
	server.loadDatabase()
	server.loadStorage()
	server.loadRepos()
	server.loadDomains()
	server.loadRouter()
	server.startServer()
}
