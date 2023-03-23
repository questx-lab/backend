package authenticator

import (
	"context"

	"golang.org/x/oauth2"
)

type IOAuth2Config interface {
	Service() string
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	VerifyIDToken(ctx context.Context, token *oauth2.Token) (string, error)
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
}

type TokenEngine[T any] interface {
	Generate(sub string, obj T) (string, error)
	Verify(token string) (T, error)
}
