package middleware

import (
	"crypto/sha256"
	"strings"

	"github.com/questx-lab/backend/internal/repository"
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

	hasedKey := sha256.Sum256([]byte(apiKey))
	owner, err := a.apiKeyRepo.GetOwnerByKey(ctx, string(hasedKey[:]))
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

	info, err := ctx.AccessTokenEngine().Verify(token)
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

	cookie, err := ctx.Request().Cookie(ctx.Configs().Auth.AccessTokenName)
	if err != nil {
		return ""
	}

	return cookie.Value
}
