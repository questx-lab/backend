package middleware

import (
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
	userRepo       repository.UserRepository
}

func NewAuthVerifier(userRepo repository.UserRepository) *AuthVerifier {
	return &AuthVerifier{userRepo: userRepo}
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
		var requestUserID string
		if a.useAccessToken {
			tokenID := verifyAccessToken(ctx)
			if tokenID != "" {
				requestUserID = tokenID
			}
		}

		if a.useAPIKey {
			projectOwnerID := a.verifyAPIKey(ctx)
			if projectOwnerID != "" {
				requestUserID = projectOwnerID
			}
		}

		if requestUserID != "" {
			if ctx.Request().URL.Path != "/updateUser" {
				user, err := a.userRepo.GetByID(ctx, requestUserID)
				if err != nil {
					ctx.Logger().Errorf("Cannot get user: %v", err)
					return errorx.Unknown
				}

				if !user
			}

			xcontext.SetRequestUserID(ctx, requestUserID)
		}

		return errorx.New(errorx.Unauthenticated, "You need to authenticate before")
	}
}

func (a *AuthVerifier) verifyAPIKey(ctx xcontext.Context) string {
	apiKey := ctx.Request().Header.Get("X-Api-Key")
	if apiKey == "" {
		return ""
	}

	owner, err := a.apiKeyRepo.GetOwnerByKey(ctx, crypto.SHA256([]byte(apiKey)))
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
	if err != nil {
		return ""
	}

	return cookie.Value
}
