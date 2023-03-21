package domain

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/api"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/authenticator"
	"github.com/questx-lab/backend/pkg/jwt"

	"github.com/gorilla/sessions"
)

type OAuth2Domain interface {
	Login(api.Context, *model.OAuth2LoginRequest) (*model.OAuth2LoginResponse, error)
	Callback(api.Context, *model.OAuth2CallbackRequest) (*model.OAuth2CallbackResponse, error)
}

type oauth2Domain struct {
	userRepo        repository.UserRepository
	oauth2Repo      repository.OAuth2Repository
	store           *sessions.CookieStore
	authenticators  []authenticator.OAuth2
	jwtEngine       *jwt.Engine[model.AccessToken]
	accessTokenName string
}

func NewOAuth2Domain(
	userRepo repository.UserRepository,
	oauth2Repo repository.OAuth2Repository,
	authenticators []authenticator.OAuth2,
	cfg config.AuthConfigs,
) OAuth2Domain {
	return &oauth2Domain{
		userRepo:        userRepo,
		oauth2Repo:      oauth2Repo,
		authenticators:  authenticators,
		store:           sessions.NewCookieStore([]byte(cfg.SessionSecret)),
		jwtEngine:       jwt.NewEngine[model.AccessToken](cfg.TokenSecret, cfg.TokenExpiration),
		accessTokenName: cfg.AccessTokenName,
	}
}

func (d *oauth2Domain) Login(
	ctx api.Context, req *model.OAuth2LoginRequest,
) (*model.OAuth2LoginResponse, error) {
	authenticator, ok := d.getAuthenticator(req.Type)
	if !ok {
		return nil, errors.New("invalid oauth2 type")
	}

	state, err := generateRandomString()
	if err != nil {
		log.Println("Cannot generate random state, err = ", err)
		return nil, errors.New("cannot generate random state")
	}

	session, err := d.store.Get(ctx.Request, authSessionKey)
	if err != nil {
		log.Println("Cannot get the session, err = ", err)
		return nil, errors.New("cannot get the session")
	}

	// Save the state inside the session.
	session.Values["state"] = state
	if err := session.Save(ctx.Request, ctx.Writer); err != nil {
		log.Println("Cannot save the session, err = ", err)
		return nil, errors.New("cannot save the session")
	}

	http.Redirect(ctx.Writer, ctx.Request,
		authenticator.AuthCodeURL(state), http.StatusTemporaryRedirect)
	return nil, nil
}

func (d *oauth2Domain) Callback(
	ctx api.Context, req *model.OAuth2CallbackRequest,
) (*model.OAuth2CallbackResponse, error) {
	auth, ok := d.getAuthenticator(req.Type)
	if !ok {
		return nil, errors.New("invalid oauth2 type")
	}

	session, err := d.store.Get(ctx.Request, authSessionKey)
	if err != nil {
		log.Println("Cannot get the session, err = ", err)
		return nil, errors.New("cannot get the session")
	}

	if state, ok := session.Values["state"]; !ok || req.State != state {
		return nil, errors.New("mismatched state parameter")
	}

	session.Options.MaxAge = -1
	if err := session.Save(ctx.Request, ctx.Writer); err != nil {
		log.Println("Cannot save the session, err = ", err)
		return nil, errors.New("cannot save the session")
	}

	// Exchange an authorization code for a serviceToken.
	serviceToken, err := auth.Exchange(ctx, req.Code)
	if err != nil {
		log.Println("Failed to exchange an authorization code for a token, err = ", err)
		return nil, errors.New("cannot exchange authorization code")
	}

	serviceID, err := auth.VerifyIDToken(ctx, serviceToken)
	if err != nil {
		log.Println("Failed to verify ID Token, err = ", err)
		return nil, errors.New("failed to verify id token")
	}

	user, err := d.userRepo.RetrieveByServiceID(ctx, auth.Name, serviceID)
	if err != nil {
		user = &entity.User{
			ID:      uuid.New().String(),
			Address: "",
			Name:    serviceID,
		}

		err = d.userRepo.Create(ctx, user)
		if err != nil {
			log.Println("Failed to create user, err = ", err)
			return nil, errors.New("cannot create a new user")
		}

		err = d.oauth2Repo.Create(ctx, &entity.OAuth2{
			UserID:        user.ID,
			Service:       auth.Name,
			ServiceUserID: serviceID,
		})
		if err != nil {
			log.Println("Failed to link user with oauth2 service, err = ", err)
			return nil, errors.New("cannot link user with oauth2 service")
		}
	}

	token, err := d.jwtEngine.Generate(user.ID, model.AccessToken{
		ID:      user.ID,
		Name:    user.Name,
		Address: user.Address,
	})
	if err != nil {
		log.Println("Failed to generate access token, err = ", err)
		return nil, errors.New("cannot generate access token")
	}

	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     d.accessTokenName,
		Value:    token,
		Domain:   "",
		Path:     "/",
		Expires:  time.Now().Add(d.jwtEngine.Expiration),
		Secure:   true,
		HttpOnly: false,
	})

	http.Redirect(ctx.Writer, ctx.Request, "/home.html", http.StatusPermanentRedirect)
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
