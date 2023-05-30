package authenticator

import (
	"context"
)

type IOAuth2Service interface {
	Service() string
	GetUserID(ctx context.Context, accessToken string) (string, error)
	VerifyIDToken(ctx context.Context, rawIDToken string) (string, error)
	VerifyAuthorizationCode(ctx context.Context, code, codeVerifier, redirectURI string) (string, error)
}
