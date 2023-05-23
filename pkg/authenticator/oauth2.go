package authenticator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/questx-lab/backend/config"
)

type oauth2Service struct {
	name        string
	verifierURL string
	idField     string

	clientID string
	provider *oidc.Provider
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
		name:        cfg.Name,
		verifierURL: cfg.VerifyURL,
		idField:     cfg.IDField,
		provider:    provider,
		clientID:    cfg.ClientID,
	}
}

func (s *oauth2Service) Service() string {
	return s.name
}

func (s *oauth2Service) GetUserID(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, s.verifierURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	info := map[string]any{}
	err = json.Unmarshal(b, &info)
	if err != nil {
		return "", err
	}

	// If idfield is foo.bar.id, user id is get from info[foo][bar][id].
	var value any = info
	for _, field := range strings.Split(s.idField, ".") {
		m, ok := value.(map[string]any)
		if !ok {
			return "", fmt.Errorf("invalid field %s in user info response", s.idField)
		}

		value, ok = m[field]
		if !ok {
			return "", fmt.Errorf("no field %s in user info response", s.idField)
		}
	}

	id, ok := value.(string)
	if !ok {
		return "", errors.New("invalid type of id field")
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

	return id, nil
}
