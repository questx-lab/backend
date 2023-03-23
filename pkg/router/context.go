package router

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/jwt"
)

type GinContext interface {
	GetHeader(key string) string
}

type Context struct {
	*gin.Context

	AccessTokenEngine *jwt.Engine[model.AccessToken]
	Configs           config.Configs
}

func (ctx *Context) GetUserID() string {
	if token := ctx.getAccessToken(); token != "" {
		if info, err := ctx.AccessTokenEngine.Verify(token); err == nil {
			return info.ID
		}
	}

	return ""
}

func (ctx *Context) getAccessToken() string {
	authorization := ctx.GetHeader("Authorization")
	auth, token, found := strings.Cut(authorization, " ")
	if found {
		if auth == "Bearer" {
			return token
		}
		return ""
	}

	cookie, err := ctx.Cookie(ctx.Configs.Auth.AccessTokenName)
	if err != nil {
		return ""
	}

	return cookie
}
