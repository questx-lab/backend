package api

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/questx-lab/backend/utils/token"
)

func Logger(ctx *Context) error {
	f, err := os.OpenFile("logs/logs.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}

	log.SetOutput(f)
	ctx.closers = append(ctx.closers, f)
	return nil
}

func ImportUserIDToContext(tknGenerator token.Generator) Handler {
	return func(ctx *Context) error {
		reqToken, err := ctx.r.Cookie(AuthCookie)
		if err != nil {
			return err
		}

		userID, err := tknGenerator.Verify(reqToken.Value)
		if err != nil {
			return err
		}
		ctx.Context = context.WithValue(ctx.Context, userCtxKey{}, userID)

		return nil
	}
}

func Close(ctx *Context) error {
	for _, closer := range ctx.closers {
		if err := closer.Close(); err != nil {
			return err
		}
	}
	return nil
}
