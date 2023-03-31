package testutil

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"golang.org/x/oauth2"
)

type mockOAuth2 struct {
	Name              string
	ExchangeFunc      func(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	VerifyIDTokenFunc func(ctx context.Context, token *oauth2.Token) (string, error)
	AuthCodeURLFunc   func(state string, opts ...oauth2.AuthCodeOption) string
}

func NewMockOAuth2() *mockOAuth2 {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	return &mockOAuth2{Name: base64.StdEncoding.EncodeToString(b)}
}

func (m *mockOAuth2) Service() string {
	return m.Name
}

func (m *mockOAuth2) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	if m.ExchangeFunc != nil {
		return m.ExchangeFunc(ctx, code, opts...)
	}

	return nil, nil
}

func (m *mockOAuth2) VerifyIDToken(ctx context.Context, token *oauth2.Token) (string, error) {
	if m.VerifyIDTokenFunc != nil {
		return m.VerifyIDTokenFunc(ctx, token)
	}

	return "", nil
}

func (m *mockOAuth2) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	if m.AuthCodeURLFunc != nil {
		return m.AuthCodeURLFunc(state, opts...)
	}

	return ""
}
