package domain

import (
	"errors"
	"fmt"
	"log"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/authenticator"
	"github.com/questx-lab/backend/pkg/router"

	"github.com/google/uuid"
)

type OAuth2Domain interface {
	Login(router.Context, *model.OAuth2LoginRequest) (*model.OAuth2LoginResponse, error)
	Callback(router.Context, *model.OAuth2CallbackRequest) (*model.OAuth2CallbackResponse, error)
}

type oauth2Domain struct {
	userRepo      repository.UserRepository
	oauth2Repo    repository.OAuth2Repository
	oauth2Configs []authenticator.IOAuth2Config
}

func NewOAuth2Domain(
	userRepo repository.UserRepository,
	oauth2Repo repository.OAuth2Repository,
	authenticators []authenticator.IOAuth2Config,
) OAuth2Domain {
	return &oauth2Domain{
		userRepo:      userRepo,
		oauth2Repo:    oauth2Repo,
		oauth2Configs: authenticators,
	}
}

func (d *oauth2Domain) Login(
	ctx router.Context, req *model.OAuth2LoginRequest,
) (*model.OAuth2LoginResponse, error) {
	authenticator, ok := d.getAuthenticator(req.Type)
	if !ok {
		return nil, fmt.Errorf("invalid oauth2 type")
	}

	state, err := generateRandomString()
	if err != nil {
		return nil, fmt.Errorf("cannot generate random string: %w", err)
	}

	return &model.OAuth2LoginResponse{
		RedirectURL: authenticator.AuthCodeURL(state),
		State:       state,
	}, nil
}

func (d *oauth2Domain) Callback(
	ctx router.Context, req *model.OAuth2CallbackRequest,
) (*model.OAuth2CallbackResponse, error) {
	auth, ok := d.getAuthenticator(req.Type)
	if !ok {
		return nil, fmt.Errorf("invalid oauth2 type")
	}

	if req.State != req.SessionState {
		return nil, fmt.Errorf("mismatched state parameter")
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

	user, err := d.userRepo.RetrieveByServiceID(ctx, auth.Service(), serviceID)
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
			Service:       auth.Service(),
			ServiceUserID: serviceID,
		})
		if err != nil {
			return nil, fmt.Errorf("cannot link user with oauth2 service: %w", err)
		}
	}

	token, err := ctx.AccessTokenEngine().Generate(user.ID, model.AccessToken{
		ID:      user.ID,
		Name:    user.Name,
		Address: user.Address,
	})
	if err != nil {
		log.Println("Failed to generate access token, err = ", err)
		return nil, errors.New("cannot generate access token")
	}

	return &model.OAuth2CallbackResponse{
		RedirectURL: "/home.html",
		AccessToken: token,
	}, nil
}

func (d *oauth2Domain) getAuthenticator(service string) (authenticator.IOAuth2Config, bool) {
	for i := range d.oauth2Configs {
		if d.oauth2Configs[i].Service() == service {
			return d.oauth2Configs[i], true
		}
	}
	return &authenticator.OAuth2Config{}, false
}
