package main

import (
	"os"
)

var server srv

func main() {
	server.loadApp()
	if err := server.app.Run(os.Args); err != nil {
		panic(err)
	}
}
