package authenticator

import (
	"context"
	"time"

	"golang.org/x/oauth2"
)

type IOAuth2Config interface {
	Service() string
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	VerifyIDToken(ctx context.Context, token *oauth2.Token) (string, error)
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
}

type TokenEngine interface {
	// Generate creates a token string containing the obj and expiration.
	Generate(expiration time.Duration, obj any) (string, error)

	// Verify if token is invalid or expired. Then parse the obj from token to obj parameter. The
	// obj paramter must be a pointer.
	Verify(token string, obj any) error
}
