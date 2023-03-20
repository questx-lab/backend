package main

import (
	"fmt"
	"net/http"
)

var server srv

func main() {
	server.loadMux()
	server.loadRepos()
	server.loadDomains()
	server.loadControllers()
	fmt.Println("Starting server")
	http.ListenAndServe(":3333", server.mux)
}
