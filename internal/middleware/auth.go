package middleware

import (
	"context"
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
	isOptional     bool
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

func (a *AuthVerifier) WithOptional() *AuthVerifier {
	a.isOptional = true
	return a
}

func (a *AuthVerifier) Middleware() router.MiddlewareFunc {
	return func(ctx context.Context) (context.Context, error) {
		if a.useAccessToken {
			tokenID := verifyAccessToken(ctx)
			if tokenID != "" {
				return xcontext.WithRequestUserID(ctx, tokenID), nil
			}
		}

		if a.useAPIKey {
			projectOwnerID := a.verifyAPIKey(ctx)
			if projectOwnerID != "" {
				return xcontext.WithRequestUserID(ctx, projectOwnerID), nil
			}
		}

		if a.isOptional {
			return nil, nil
		} else {
			return nil, errorx.New(errorx.Unauthenticated, "You need to authenticate before")
		}
	}
}

func (a *AuthVerifier) verifyAPIKey(ctx context.Context) string {
	apiKey := xcontext.HTTPRequest(ctx).Header.Get("X-Api-Key")
	if apiKey == "" {
		return ""
	}

	owner, err := a.apiKeyRepo.GetOwnerByKey(ctx, crypto.SHA256([]byte(apiKey)))
	if err != nil {
		return ""
	}

	return owner
}

func verifyAccessToken(ctx context.Context) string {
	token := getAccessToken(ctx)
	if token == "" {
		return ""
	}

	var info model.AccessToken
	err := xcontext.TokenEngine(ctx).Verify(token, &info)
	if err != nil {
		return ""
	}

	return info.ID
}

func getAccessToken(ctx context.Context) string {
	req := xcontext.HTTPRequest(ctx)
	authorization := req.Header.Get("Authorization")
	auth, token, found := strings.Cut(authorization, " ")
	if found {
		if auth == "Bearer" {
			return token
		}
		return ""
	}

	cookie, err := req.Cookie(xcontext.Configs(ctx).Auth.AccessToken.Name)
	if err != nil {
		return ""
	}

	return cookie.Value
}
