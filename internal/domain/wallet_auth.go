package domain

import (
	"bytes"
	"errors"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/questx-lab/backend/api"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/jwt"
)

type WalletAuthDomain interface {
	Login(api.CustomContext, *model.WalletLoginRequest) (*model.WalletLoginResponse, error)
	Verify(api.CustomContext, *model.WalletVerifyRequest) (*model.WalletVerifyResponse, error)
}

type walletAuthDomain struct {
	userRepo  repository.UserRepository
	store     *sessions.CookieStore
	jwtEngine *jwt.Engine[model.AccessToken]
}

func (d *walletAuthDomain) Login(
	ctx api.CustomContext, req *model.WalletLoginRequest,
) (*model.WalletLoginResponse, error) {
	nonce, err := generateRandomString()
	if err != nil {
		log.Println("Cannot generate random state, err = ", err)
		return nil, errors.New("cannot generate random state")
	}

	session, err := d.store.Get(ctx.Request, authSessionKey)
	if err != nil {
		log.Println("Cannot get the session, err = ", err)
		return nil, errors.New("cannot get the session")
	}

	// Save nonce and address inside the session.
	session.Values["nonce"] = nonce
	session.Values["address"] = req.Address
	if err := session.Save(ctx.Request, ctx.Writer); err != nil {
		log.Println("Cannot save the session, err = ", err)
		return nil, errors.New("cannot save the session")
	}

	return &model.WalletLoginResponse{Nonce: nonce}, nil
}

func (d *walletAuthDomain) Verify(
	ctx api.CustomContext, req *model.WalletVerifyRequest,
) (*model.WalletVerifyResponse, error) {
	session, err := d.store.Get(ctx.Request, authSessionKey)
	if err != nil {
		log.Println("Cannot get the session, err = ", err)
		return nil, errors.New("cannot get the session")
	}

	nonceObj, ok := session.Values["state"]
	if !ok {
		return nil, errors.New("cannot get nonce from session")
	}

	addressObj, ok := session.Values["address"]
	if !ok {
		return nil, errors.New("cannot get address from session")
	}

	nonce := nonceObj.(string)
	address := addressObj.(string)

	session.Options.MaxAge = -1
	if err := session.Save(ctx.Request, ctx.Writer); err != nil {
		log.Println("Cannot save the session, err = ", err)
		return nil, errors.New("cannot save the session")
	}

	hash := crypto.Keccak256([]byte(nonce))
	signatureAddress, err := crypto.Ecrecover(hash, []byte(req.Signature))
	if err != nil {
		log.Println("Cannot recover the signature, err = ", err)
		return nil, errors.New("cannot recover the signature")
	}

	if bytes.Compare(signatureAddress, []byte(address)) != 0 {
		return nil, errors.New("invalid signature")
	}

	user, err := d.userRepo.RetrieveByAddress(ctx, address)
	if err != nil {
		user = &entity.User{
			ID:      uuid.New(),
			Address: address,
			Name:    address,
		}

		err = d.userRepo.Create(ctx, user)
		if err != nil {
			log.Println("Failed to create user, err = ", err)
			return nil, errors.New("cannot create a new user")
		}
	}

	token, err := d.jwtEngine.Generate(user.ID.String(), model.AccessToken{
		ID:      user.ID.String(),
		Name:    user.Name,
		Address: user.Address,
	})
	if err != nil {
		log.Println("Failed to generate access token, err = ", err)
		return nil, errors.New("cannot generate access token")
	}

	return &model.WalletVerifyResponse{AccessToken: token}, nil
}
