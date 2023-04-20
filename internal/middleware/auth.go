package middleware

import (
	"net/http"
	"strings"

	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/crypto"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type AuthVerifier struct {
	useAccessToken bool
	useAPIKey      bool
	apiKeyRepo     repository.APIKeyRepository
}

func NewAuthVerifier() *AuthVerifier {
	return &AuthVerifier{}
}

func (a *AuthVerifier) WithAccessToken() *AuthVerifier {
	a.useAccessToken = true
	return a
}

func (a *AuthVerifier) WithAPIKey(apiKeyRepo repository.APIKeyRepository) *AuthVerifier {
	a.useAPIKey = true
	a.apiKeyRepo = apiKeyRepo
	return a
}

func (a *AuthVerifier) Middleware() router.MiddlewareFunc {
	return func(ctx xcontext.Context) error {
		if a.useAccessToken {
			tokenID := verifyAccessToken(ctx)
			if tokenID != "" {
				xcontext.SetRequestUserID(ctx, tokenID)
				return nil
			}
		}

		if a.useAPIKey {
			projectOwnerID := a.verifyAPIKey(ctx)
			if projectOwnerID != "" {
				xcontext.SetRequestUserID(ctx, projectOwnerID)
				return nil
			}
		}

		return errorx.New(errorx.Unauthenticated, "You need to authenticate before")
	}
}

func (a *AuthVerifier) verifyAPIKey(ctx xcontext.Context) string {
	apiKey := ctx.Request().Header.Get("X-Api-Key")
	if apiKey == "" {
		return ""
	}

	owner, err := a.apiKeyRepo.GetOwnerByKey(ctx, crypto.Hash([]byte(apiKey)))
	if err != nil {
		return ""
	}

	return owner
}

func verifyAccessToken(ctx xcontext.Context) string {
	token := getAccessToken(ctx)
	if token == "" {
		return ""
	}

	var info model.AccessToken
	err := ctx.TokenEngine().Verify(token, &info)
	if err != nil {
		return ""
	}

	return info.ID
}

func getAccessToken(ctx xcontext.Context) string {
	authorization := ctx.Request().Header.Get("Authorization")
	auth, token, found := strings.Cut(authorization, " ")
	if found {
		if auth == "Bearer" {
			return token
		}
		return ""
	}

	cookie, err := ctx.Request().Cookie(ctx.Configs().Auth.AccessToken.Name)
	if err != http.ErrNoCookie {
		if err == nil {
			return cookie.Value
		}
		return ""
	}

	tokenFromPath, ok := ctx.Request().URL.Query()[ctx.Configs().Auth.AccessToken.Name]
	if ok && len(tokenFromPath) > 0 {
		return tokenFromPath[0]
	}

	return ""
}
