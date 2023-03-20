package main

import (
	"fmt"
	"net/http"
)

var server srv

func main() {
	fmt.Println("Starting server")
	http.ListenAndServe(":3333", server.mux)
}
