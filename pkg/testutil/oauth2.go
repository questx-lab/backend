package testutil

import (
	"context"
)

type mockOAuth2 struct {
	Name                        string
	GetUserIDFunc               func(context.Context, string) (string, error)
	VerifyIDTokenFunc           func(context.Context, string) (string, error)
	VerifyAuthorizationCodeFunc func(context.Context, string, string, string) (string, error)
}

func NewMockOAuth2(name string) *mockOAuth2 {
	return &mockOAuth2{Name: name}
}

func (m *mockOAuth2) Service() string {
	return m.Name
}

func (m *mockOAuth2) GetUserID(ctx context.Context, accessToken string) (string, error) {
	if m.GetUserIDFunc != nil {
		return m.GetUserIDFunc(ctx, accessToken)
	}
	return "", nil
}

func (m *mockOAuth2) VerifyIDToken(ctx context.Context, rawIDToken string) (string, error) {
	if m.VerifyIDTokenFunc != nil {
		return m.VerifyIDTokenFunc(ctx, rawIDToken)
	}
	return "", nil
}

func (m *mockOAuth2) VerifyAuthorizationCode(ctx context.Context, code, codeVerifier, redirectURI string) (string, error) {
	if m.VerifyAuthorizationCodeFunc != nil {
		return m.VerifyAuthorizationCodeFunc(ctx, code, codeVerifier, redirectURI)
	}
	return "", nil
}
