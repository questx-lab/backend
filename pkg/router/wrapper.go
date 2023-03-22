package router

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func wrapHandler[Request, Response any](
	router *Router,
	method string,
	handler HandlerFunc[Request, Response],
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Request
		var err error
		switch method {
		case "GET":
			err = ctx.BindQuery(&req)
		case "POST":
			err = ctx.BindJSON(&req)
		default:
			err = errors.New("unsupported method")
		}
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		resp, err := handler(&Context{
			Context:           ctx,
			AccessTokenEngine: router.accessTokenEngine,
			Configs:           router.cfg,
		}, req)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else if resp != nil {
			ctx.JSON(http.StatusOK, resp)
		}
	}
}

func wrapMiddleware(router *Router, middleware MiddlewareFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		customContext := &Context{
			Context: ctx,

			AccessTokenEngine: router.accessTokenEngine,
			Configs:           router.cfg,
		}

		err := middleware(customContext)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			ctx.Abort()
		}
	}
}
