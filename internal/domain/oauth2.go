package domain

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/authenticator"
	"github.com/questx-lab/backend/pkg/router"

	"github.com/google/uuid"
)

type OAuth2Domain interface {
	Login(*router.Context, model.OAuth2LoginRequest) (*model.OAuth2LoginResponse, error)
	Callback(*router.Context, model.OAuth2CallbackRequest) (*model.OAuth2CallbackResponse, error)
}

type oauth2Domain struct {
	userRepo       repository.UserRepository
	oauth2Repo     repository.OAuth2Repository
	authenticators []authenticator.OAuth2
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

func (d *oauth2Domain) Login(
	ctx *router.Context, req model.OAuth2LoginRequest,
) (*model.OAuth2LoginResponse, error) {
	authenticator, ok := d.getAuthenticator(req.Type)
	if !ok {
		return nil, fmt.Errorf("invalid oauth2 type")
	}

	state, err := generateRandomString()
	if err != nil {
		return nil, errors.New("cannot generate random string")
	}

	session := sessions.Default(ctx.Context)
	session.Set("state", state)
	if err := session.Save(); err != nil {
		return nil, fmt.Errorf("cannot save the session: %w", err)
	}

	ctx.Redirect(http.StatusTemporaryRedirect, authenticator.AuthCodeURL(state))
	return nil, nil
}

func (d *oauth2Domain) Callback(
	ctx *router.Context, req model.OAuth2CallbackRequest,
) (*model.OAuth2CallbackResponse, error) {
	auth, ok := d.getAuthenticator(req.Type)
	if !ok {
		return nil, fmt.Errorf("invalid oauth2 type")
	}

	session := sessions.Default(ctx.Context)
	state := session.Get("state")
	if req.State != state {
		return nil, fmt.Errorf("mismatched state parameter")
	}

	session.Delete("state")
	if err := session.Save(); err != nil {
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

	ctx.SetCookie(
		ctx.Configs.Auth.AccessTokenName, // name
		token,                            // value
		int(ctx.Configs.Token.Expiration.Seconds()), // max-age
		"",          // path
		"localhost", // domain
		true,        // secure
		false,       // httpOnly
	)

	ctx.Redirect(http.StatusTemporaryRedirect, "/static/home.html")
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
