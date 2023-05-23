package authenticator

import (
	"context"
	"time"
)

type IOAuth2Service interface {
	Service() string
	GetUserID(ctx context.Context, accessToken string) (string, error)
	VerifyIDToken(ctx context.Context, rawIDToken string) (string, error)
}

type TokenEngine interface {
	// Generate creates a token string containing the obj and expiration.
	Generate(expiration time.Duration, obj any) (string, error)

	// Verify if token is invalid or expired. Then parse the obj from token to obj parameter. The
	// obj paramter must be a pointer.
	Verify(token string, obj any) error
}
