package domain

import (
	"bytes"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-contrib/sessions"
	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/router"
)

type WalletAuthDomain interface {
	Login(*router.Context, model.WalletLoginRequest) (*model.WalletLoginResponse, error)
	Verify(*router.Context, model.WalletVerifyRequest) (*model.WalletVerifyResponse, error)
}

type walletAuthDomain struct {
	userRepo repository.UserRepository
}

func NewWalletAuthDomain(userRepo repository.UserRepository) WalletAuthDomain {
	return &walletAuthDomain{userRepo: userRepo}
}

func (d *walletAuthDomain) Login(
	ctx *router.Context, req model.WalletLoginRequest,
) (*model.WalletLoginResponse, error) {
	nonce, err := generateRandomString()
	if err != nil {
		return nil, fmt.Errorf("cannot generate random state: %w", err)
	}

	session := sessions.Default(ctx.Context)

	// Save nonce and address inside the session.
	session.Set("nonce", nonce)
	session.Set("address", req.Address)
	if err := session.Save(); err != nil {
		return nil, fmt.Errorf("cannot save the session: %w", err)
	}

	return &model.WalletLoginResponse{Nonce: nonce}, nil
}

func (d *walletAuthDomain) Verify(
	ctx *router.Context, req model.WalletVerifyRequest,
) (*model.WalletVerifyResponse, error) {
	session := sessions.Default(ctx.Context)

	nonce, ok := session.Get("nonce").(string)
	if !ok {
		return nil, fmt.Errorf("cannot get nonce from session")
	}

	address, ok := session.Get("address").(string)
	if !ok {
		return nil, fmt.Errorf("cannot get address from session")
	}

	session.Delete("nonce")
	session.Delete("address")
	if err := session.Save(); err != nil {
		return nil, fmt.Errorf("cannot save the session: %w", err)
	}

	hash := accounts.TextHash([]byte(nonce))
	signature, err := hexutil.Decode(req.Signature)
	if err != nil {
		return nil, fmt.Errorf("cannot decode the signature: %w", err)
	}

	if signature[crypto.RecoveryIDOffset] == 27 || signature[crypto.RecoveryIDOffset] == 28 {
		signature[crypto.RecoveryIDOffset] -= 27 // Transform yellow paper V from 27/28 to 0/1
	}

	recovered, err := crypto.SigToPub(hash, signature)
	if err != nil {
		log.Println("Cannot recover signature, err = ", err)
		return nil, fmt.Errorf("cannot recover signature: %w", err)
	}

	recoveredAddr := crypto.PubkeyToAddress(*recovered)
	if !bytes.Equal(recoveredAddr.Bytes(), common.HexToAddress(address).Bytes()) {
		return nil, fmt.Errorf("mismatched address")
	}

	user, err := d.userRepo.RetrieveByAddress(ctx, address)
	if err != nil {
		user = &entity.User{
			Base:    entity.Base{ID: uuid.NewString()},
			Address: address,
			Name:    address,
		}

		err = d.userRepo.Create(ctx, user)
		if err != nil {
			return nil, fmt.Errorf("cannot create a new user: %w", err)
		}
	}

	token, err := ctx.AccessTokenEngine.Generate(user.ID, model.AccessToken{
		ID:      user.ID,
		Name:    user.Name,
		Address: user.Address,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot generate access token: %w", err)
	}

	ctx.SetCookie(
		ctx.Configs.Auth.AccessTokenName, // name
		token,                            // value
		int(ctx.Configs.Token.Expiration.Seconds()), // max-age
		"/",   // path
		"",    // domain
		true,  // secure
		false, // httpOnly
	)

	return &model.WalletVerifyResponse{AccessToken: token}, nil
}
