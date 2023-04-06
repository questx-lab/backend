package testutil

import (
	"context"
)

type mockOAuth2 struct {
	Name          string
	GetUserIDFunc func(context.Context, string) (string, error)
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
