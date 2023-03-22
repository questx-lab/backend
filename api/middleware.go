package api

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/questx-lab/backend/utils/token"
)

func Logger(ctx *Context) {
	f, err := os.OpenFile("logs/logs.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	log.SetOutput(f)
	ctx.closers = append(ctx.closers, f)
}

func ImportUserIDToContext(tknGenerator token.Generator) Handler {
	return func(ctx *Context) {
		reqToken, err := ctx.r.Cookie(AuthCookie)
		if err != nil {
			http.Error(ctx.w, err.Error(), http.StatusBadRequest)
			return
		}

		userID, err := tknGenerator.Verify(reqToken.Value)
		if err != nil {
			http.Error(ctx.w, err.Error(), http.StatusBadRequest)
			return
		}
		ctx.Context = context.WithValue(ctx.Context, userCtxKey{}, userID)
	}
}

func Close(ctx *Context) {
	for _, closer := range ctx.closers {
		closer.Close()
	}
}
