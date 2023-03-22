package api

import (
	"fmt"
	"log"
	"os"
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

func Close(ctx *Context) error {
	for _, closer := range ctx.closers {
		if err := closer.Close(); err != nil {
			return err
		}
	}
	return nil
}
