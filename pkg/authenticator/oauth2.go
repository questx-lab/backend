package authenticator

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type OAuth2 struct {
	*oidc.Provider
	oauth2.Config

	Name    string
	idField string
}

func NewOAuth2(
	ctx context.Context, service, issuer, clientID, clientSecret string, idField string,
) (OAuth2, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return OAuth2{}, err
	}

	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  "https://localhost:8080/oauth2/callback?type=" + service,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
	}

	return OAuth2{Name: service, Provider: provider, Config: config}, nil
}

// VerifyIDToken verifies that an *oauth2.Token is a valid *oidc.IDToken.
func (a OAuth2) VerifyIDToken(ctx context.Context, token *oauth2.Token) (string, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return "", errors.New("no id_token field in oauth2 token")
	}

	oidcConfig := &oidc.Config{
		ClientID: a.ClientID,
	}

	idToken, err := a.Verifier(oidcConfig).Verify(ctx, rawIDToken)
	if err != nil {
		return "", err
	}

	var profile map[string]interface{}
	if err = idToken.Claims(&profile); err != nil {
		return "", errors.New("invalid id token")
	}

	id, ok := profile[a.idField].(string)
	if !ok {
		return "", fmt.Errorf("invalid id field %s", a.idField)
	}

	return id, nil
}
