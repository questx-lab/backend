package authenticator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/questx-lab/backend/config"
)

type oauth2Service struct {
	name        string
	verifierURL string
	idField     string
}

func NewOAuth2Service(cfg config.OAuth2Config) *oauth2Service {
	return &oauth2Service{name: cfg.Name, verifierURL: cfg.VerifyURL, idField: cfg.IDField}
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
