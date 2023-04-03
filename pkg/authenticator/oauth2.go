package authenticator

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/questx-lab/backend/config"
	"golang.org/x/oauth2"
)

type OAuth2Config struct {
	*oidc.Provider
	oauth2.Config

	name    string
	idField string
}

func NewOAuth2Config(
	ctx context.Context, cfg config.Configs, oauth2Cfg config.OAuth2Config,
) (IOAuth2Config, error) {
	provider, err := oidc.NewProvider(ctx, oauth2Cfg.Issuer)
	if err != nil {
		return &OAuth2Config{}, err
	}

	config := oauth2.Config{
		ClientID:     oauth2Cfg.ClientID,
		ClientSecret: oauth2Cfg.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL: fmt.Sprintf("http://%s:%s/oauth2/callback?type=%s",
			cfg.Server.Host, cfg.Server.Port, oauth2Cfg.Name),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
	}

	return &OAuth2Config{name: oauth2Cfg.Name, idField: oauth2Cfg.IDField, Provider: provider, Config: config}, nil
}

func (a *OAuth2Config) Service() string {
	return a.name
}

// VerifyIDToken verifies that an *oauth2.Token is a valid *oidc.IDToken.
func (a *OAuth2Config) VerifyIDToken(ctx context.Context, token *oauth2.Token) (string, error) {
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
