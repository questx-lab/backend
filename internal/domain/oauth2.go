package domain

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/api"
	"github.com/questx-lab/backend/pkg/authenticator"

	"github.com/google/uuid"
)

type OAuth2Domain interface {
	Login(*api.Context, *model.OAuth2LoginRequest) (*model.OAuth2LoginResponse, error)
	Callback(*api.Context, *model.OAuth2CallbackRequest) (*model.OAuth2CallbackResponse, error)
}

type oauth2Domain struct {
	userRepo       repository.UserRepository
	oauth2Repo     repository.OAuth2Repository
	authenticators []authenticator.OAuth2
	cfg            config.AuthConfigs
}

func NewOAuth2Domain(
	userRepo repository.UserRepository,
	oauth2Repo repository.OAuth2Repository,
	authenticators []authenticator.OAuth2,
) OAuth2Domain {
	return &oauth2Domain{
		userRepo:       userRepo,
		oauth2Repo:     oauth2Repo,
		authenticators: authenticators,
	}
}

func (d *oauth2Domain) Login(ctx *api.Context, req *model.OAuth2LoginRequest) (*model.OAuth2LoginResponse, error) {
	r := ctx.GetRequest()
	w := ctx.GetResponse()
	authenticator, ok := d.getAuthenticator(req.Type)
	if !ok {
		return nil, fmt.Errorf("invalid oauth2 type")
	}

	state := uuid.NewString()

	session, err := ctx.SessionStore.Get(r)
	if err != nil {
		return nil, fmt.Errorf("cannot get the session: %w", err)
	}

	// Save the state inside the session.
	session.Values["state"] = state
	if err := session.Save(r, w); err != nil {
		return nil, fmt.Errorf("cannot save the session: %w", err)
	}

	http.Redirect(w, r,
		authenticator.AuthCodeURL(state), http.StatusTemporaryRedirect)
	return nil, nil
}

func (d *oauth2Domain) Callback(ctx *api.Context, req *model.OAuth2CallbackRequest) (*model.OAuth2CallbackResponse, error) {
	r := ctx.GetRequest()
	w := ctx.GetResponse()
	auth, ok := d.getAuthenticator(req.Type)
	if !ok {
		return nil, fmt.Errorf("invalid oauth2 type")
	}

	session, err := ctx.SessionStore.Get(r)
	if err != nil {
		return nil, fmt.Errorf("cannot get the session: %w", err)
	}

	if state, ok := session.Values["state"]; !ok || req.State != state {
		return nil, fmt.Errorf("mismatched state parameter")
	}

	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		return nil, fmt.Errorf("cannot save the session: %w", err)
	}

	// Exchange an authorization code for a serviceToken.
	serviceToken, err := auth.Exchange(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("cannot exchange authorization code: %w", err)
	}

	serviceID, err := auth.VerifyIDToken(ctx, serviceToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify id token: %w", err)
	}

	user, err := d.userRepo.RetrieveByServiceID(ctx, auth.Name, serviceID)
	if err != nil {
		user = &entity.User{
			Base:    entity.Base{ID: uuid.NewString()},
			Address: "",
			Name:    serviceID,
		}

		err = d.userRepo.Create(ctx, user)
		if err != nil {
			return nil, fmt.Errorf("cannot create a new user: %w", err)
		}

		err = d.oauth2Repo.Create(ctx, &entity.OAuth2{
			UserID:        user.ID,
			Service:       auth.Name,
			ServiceUserID: serviceID,
		})
		if err != nil {
			return nil, fmt.Errorf("cannot link user with oauth2 service: %w", err)
		}
	}

	token, err := ctx.AccessTokenEngine.Generate(user.ID, model.AccessToken{
		ID:      user.ID,
		Name:    user.Name,
		Address: user.Address,
	})
	if err != nil {
		log.Println("Failed to generate access token, err = ", err)
		return nil, errors.New("cannot generate access token")
	}

	http.SetCookie(w, &http.Cookie{
		Name:     ctx.Configs.Auth.AccessTokenName,
		Value:    token,
		Domain:   "",
		Path:     "/",
		Expires:  time.Now().Add(ctx.Configs.Token.Expiration),
		Secure:   true,
		HttpOnly: false,
	})

	http.Redirect(w, r, "/home.html", http.StatusPermanentRedirect)
	return nil, nil
}

func (d *oauth2Domain) getAuthenticator(service string) (authenticator.OAuth2, bool) {
	for i := range d.authenticators {
		if d.authenticators[i].Name == service {
			return d.authenticators[i], true
		}
	}
	return authenticator.OAuth2{}, false
}
