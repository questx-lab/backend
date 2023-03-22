package domain

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/questx-lab/backend/api"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/jwt"
)

type WalletAuthDomain interface {
	Login(*api.Context, *model.WalletLoginRequest) (*model.WalletLoginResponse, error)
	Verify(*api.Context, *model.WalletVerifyRequest) (*model.WalletVerifyResponse, error)
}

type walletAuthDomain struct {
	userRepo        repository.UserRepository
	store           *sessions.CookieStore
	jwtEngine       *jwt.Engine[model.AccessToken]
	accessTokenName string
}

func NewWalletAuthDomain(
	userRepo repository.UserRepository,
	cfg config.AuthConfigs,
) WalletAuthDomain {
	return &walletAuthDomain{
		userRepo:        userRepo,
		store:           sessions.NewCookieStore([]byte(cfg.SessionSecret)),
		jwtEngine:       jwt.NewEngine[model.AccessToken](cfg.TokenSecret, cfg.TokenExpiration),
		accessTokenName: cfg.AccessTokenName,
	}
}

func (d *walletAuthDomain) Login(
	ctx *api.Context, req *model.WalletLoginRequest,
) (*model.WalletLoginResponse, error) {
	r := ctx.GetRequest()
	w := ctx.GetResponse()
	nonce, err := generateRandomString()
	if err != nil {
		log.Println("Cannot generate random state, err = ", err)
		return nil, errors.New("cannot generate random state")
	}

	session, err := d.store.Get(r, authSessionKey)
	if err != nil {
		log.Println("Cannot get the session, err = ", err)
		return nil, errors.New("cannot get the session")
	}

	// Save nonce and address inside the session.
	session.Values["nonce"] = nonce
	session.Values["address"] = req.Address
	if err := session.Save(r, w); err != nil {
		log.Println("Cannot save the session, err = ", err)
		return nil, errors.New("cannot save the session")
	}

	return &model.WalletLoginResponse{Nonce: nonce}, nil
}

func (d *walletAuthDomain) Verify(
	ctx *api.Context, req *model.WalletVerifyRequest,
) (*model.WalletVerifyResponse, error) {
	r := ctx.GetRequest()
	w := ctx.GetResponse()
	session, err := d.store.Get(r, authSessionKey)
	if err != nil {
		log.Println("Cannot get the session, err = ", err)
		return nil, errors.New("cannot get the session")
	}

	nonceObj, ok := session.Values["nonce"]
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
	if err := session.Save(r, w); err != nil {
		log.Println("Cannot save the session, err = ", err)
		return nil, errors.New("cannot save the session")
	}

	hash := accounts.TextHash([]byte(nonce))
	signature, err := hexutil.Decode(req.Signature)
	if err != nil {
		log.Println("Cannot decode the signature, err = ", err)
		return nil, errors.New("cannot decode the signature")
	}

	if signature[crypto.RecoveryIDOffset] == 27 || signature[crypto.RecoveryIDOffset] == 28 {
		signature[crypto.RecoveryIDOffset] -= 27 // Transform yellow paper V from 27/28 to 0/1
	}

	recovered, err := crypto.SigToPub(hash, signature)
	if err != nil {
		log.Println("Cannot recover signature, err = ", err)
		return nil, errors.New("cannot recover signature")
	}

	recoveredAddr := crypto.PubkeyToAddress(*recovered)
	if bytes.Compare(recoveredAddr.Bytes(), common.HexToAddress(address).Bytes()) != 0 {
		return nil, errors.New("mismatched address")
	}

	user, err := d.userRepo.RetrieveByAddress(ctx, address)
	if err != nil {
		user = &entity.User{
			Base: entity.Base{
				ID: uuid.New().String(),
			},
			Address: address,
			Name:    address,
		}

		err = d.userRepo.Create(ctx, user)
		if err != nil {
			log.Println("Failed to create user, err = ", err)
			return nil, errors.New("cannot create a new user")
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

	http.SetCookie(w, &http.Cookie{
		Name:     d.accessTokenName,
		Value:    token,
		Domain:   "",
		Path:     "/",
		Expires:  time.Now().Add(d.jwtEngine.Expiration),
		Secure:   true,
		HttpOnly: false,
	})

	return &model.WalletVerifyResponse{AccessToken: token}, nil
}
