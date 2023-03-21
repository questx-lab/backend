package api

import (
	"context"
	"net/http"
)

type CustomContext struct {
	context.Context

	Request *http.Request
	Writer  http.ResponseWriter
}
