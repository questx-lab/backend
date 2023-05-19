package main

import (
	"context"
	"net/http"
	"os"

	"github.com/questx-lab/backend/pkg/xcontext"
)

var server srv

func main() {
	server.ctx = context.Background()
	server.ctx = xcontext.WithHTTPClient(server.ctx, http.DefaultClient)
	server.loadApp()
	if err := server.app.Run(os.Args); err != nil {
		panic(err)
	}
}
