package authenticator

import (
	"context"
)

type OAuth2User struct {
	ID       string
	Username string
}

type IOAuth2Service interface {
	Service() string
	GetUserID(ctx context.Context, accessToken string) (OAuth2User, error)
	VerifyIDToken(ctx context.Context, rawIDToken string) (OAuth2User, error)
	VerifyAuthorizationCode(ctx context.Context, code, codeVerifier, redirectURI string) (OAuth2User, error)
}
