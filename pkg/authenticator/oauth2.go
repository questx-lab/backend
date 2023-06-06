package authenticator

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/pkg/api"
)

type oauth2Service struct {
	name        string
	verifierURL string
	idField     string

	clientID     string
	provider     *oidc.Provider
	tokenURL     string
	apiGenerator api.Generator
}

func NewOAuth2Service(ctx context.Context, cfg config.OAuth2Config) *oauth2Service {
	var provider *oidc.Provider
	if cfg.Issuer != "" {
		var err error
		provider, err = oidc.NewProvider(ctx, cfg.Issuer)
		if err != nil {
			panic(err)
		}
	}

	return &oauth2Service{
		name:         cfg.Name,
		verifierURL:  cfg.VerifyURL,
		idField:      cfg.IDField,
		provider:     provider,
		tokenURL:     cfg.TokenURL,
		clientID:     cfg.ClientID,
		apiGenerator: api.NewGenerator(),
	}
}

func (s *oauth2Service) Service() string {
	return s.name
}

func (s *oauth2Service) GetUserID(ctx context.Context, accessToken string) (string, error) {
	resp, err := s.apiGenerator.New(s.verifierURL, "").
		GET(ctx, api.OAuth2("Bearer", accessToken))

	if err != nil {
		return "", err
	}

	if resp.Code != 200 {
		return "", fmt.Errorf("invalid status code: %d", resp.Code)
	}

	body, ok := resp.Body.(api.JSON)
	if !ok {
		return "", errors.New("invalid body format")
	}

	id, err := body.GetString(s.idField)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s_%s", s.Service(), id), nil
}

// VerifyIDToken verifies a raw idtoken.
func (s *oauth2Service) VerifyIDToken(ctx context.Context, rawIDToken string) (string, error) {
	if s.provider == nil {
		return "", fmt.Errorf("not setting up verify idtoken feature for %s", s.name)
	}

	oidcConfig := &oidc.Config{ClientID: s.clientID}
	idToken, err := s.provider.Verifier(oidcConfig).Verify(ctx, rawIDToken)
	if err != nil {
		return "", err
	}

	var profile map[string]interface{}
	if err = idToken.Claims(&profile); err != nil {
		return "", errors.New("invalid id token")
	}

	id, ok := profile[s.idField].(string)
	if !ok {
		return "", fmt.Errorf("invalid id field %s", s.idField)
	}

	return fmt.Sprintf("%s_%s", s.name, id), nil
}

func (s *oauth2Service) VerifyAuthorizationCode(
	ctx context.Context, code, codeVerifier, redirectURI string,
) (string, error) {
	tokenURL := s.tokenURL
	if s.provider != nil {
		tokenURL = s.provider.Endpoint().TokenURL
	}

	if tokenURL == "" {
		return "", fmt.Errorf("not support authorization code verification of %s", s.name)
	}

	resp, err := s.apiGenerator.New(tokenURL, "").
		Body(api.Parameter{
			"code":          code,
			"code_verifier": codeVerifier,
			"redirect_uri":  redirectURI,
			"grant_type":    "authorization_code",
			"client_id":     s.clientID,
		}).
		POST(ctx)
	if err != nil {
		return "", err
	}

	body, ok := resp.Body.(api.JSON)
	if !ok {
		return "", errors.New("invalid body format")
	}

	if resp.Code != 200 {
		errorDesc, err := body.GetString("error_description")
		if err != nil {
			return "", fmt.Errorf("invalid status code (%d)", resp.Code)
		}

		return "", fmt.Errorf("invalid status code (%d): %s", resp.Code, errorDesc)
	}

	accessToken, err := body.GetString("access_token")
	if err != nil {
		return "", err
	}

	return s.GetUserID(ctx, accessToken)
}
