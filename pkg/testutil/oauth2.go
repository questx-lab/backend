package testutil

import (
	"context"

	"github.com/questx-lab/backend/pkg/authenticator"
)

type mockOAuth2 struct {
	Name                        string
	GetUserIDFunc               func(context.Context, string) (authenticator.OAuth2User, error)
	VerifyIDTokenFunc           func(context.Context, string) (authenticator.OAuth2User, error)
	VerifyAuthorizationCodeFunc func(context.Context, string, string, string) (authenticator.OAuth2User, error)
}

func NewMockOAuth2(name string) *mockOAuth2 {
	return &mockOAuth2{Name: name}
}

func (m *mockOAuth2) Service() string {
	return m.Name
}

func (m *mockOAuth2) GetUserID(ctx context.Context, accessToken string) (authenticator.OAuth2User, error) {
	if m.GetUserIDFunc != nil {
		return m.GetUserIDFunc(ctx, accessToken)
	}
	return authenticator.OAuth2User{}, nil
}

func (m *mockOAuth2) VerifyIDToken(ctx context.Context, rawIDToken string) (authenticator.OAuth2User, error) {
	if m.VerifyIDTokenFunc != nil {
		return m.VerifyIDTokenFunc(ctx, rawIDToken)
	}
	return authenticator.OAuth2User{}, nil
}

func (m *mockOAuth2) VerifyAuthorizationCode(
	ctx context.Context, code, codeVerifier, redirectURI string,
) (authenticator.OAuth2User, error) {
	if m.VerifyAuthorizationCodeFunc != nil {
		return m.VerifyAuthorizationCodeFunc(ctx, code, codeVerifier, redirectURI)
	}
	return authenticator.OAuth2User{}, nil
}
