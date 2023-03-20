package api

import (
	"context"
	"net/http"
)

type CustomContext struct {
	r   *http.Request
	w   http.ResponseWriter
	ctx context.Context
}
