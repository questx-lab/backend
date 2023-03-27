package domain

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/router"
)

type WalletAuthDomain interface {
	Login(router.Context, *model.WalletLoginRequest) (*model.WalletLoginResponse, error)
	Verify(router.Context, *model.WalletVerifyRequest) (*model.WalletVerifyResponse, error)
}

type walletAuthDomain struct {
	userRepo repository.UserRepository
}

func NewWalletAuthDomain(userRepo repository.UserRepository) WalletAuthDomain {
	return &walletAuthDomain{userRepo: userRepo}
}

func (d *walletAuthDomain) Login(
	ctx router.Context, req *model.WalletLoginRequest,
) (*model.WalletLoginResponse, error) {
	nonce, err := generateRandomString()
	if err != nil {
		return nil, errorx.NewGeneric(err, "Cannot generate state")
	}

	return &model.WalletLoginResponse{Address: req.Address, Nonce: nonce}, nil
}

func (d *walletAuthDomain) Verify(
	ctx router.Context, req *model.WalletVerifyRequest,
) (*model.WalletVerifyResponse, error) {
	hash := accounts.TextHash([]byte(req.SessionNonce))
	signature, err := hexutil.Decode(req.Signature)
	if err != nil {
		return nil, errorx.NewGeneric(err, "Cannot decode signature")
	}

	if signature[crypto.RecoveryIDOffset] == 27 || signature[crypto.RecoveryIDOffset] == 28 {
		signature[crypto.RecoveryIDOffset] -= 27 // Transform yellow paper V from 27/28 to 0/1
	}

	recovered, err := crypto.SigToPub(hash, signature)
	if err != nil {
		return nil, errorx.NewGeneric(err, "Cannot recover signature to public key")
	}

	recoveredAddr := crypto.PubkeyToAddress(*recovered)
	if !bytes.Equal(recoveredAddr.Bytes(), common.HexToAddress(req.SessionAddress).Bytes()) {
		return nil, errorx.NewGeneric(
			fmt.Errorf("%s != %s", recoveredAddr, req.SessionAddress), "Mismatched address")
	}

	user, err := d.userRepo.GetByAddress(ctx, req.SessionAddress)
	if err != nil {
		user = &entity.User{
			Base:    entity.Base{ID: uuid.NewString()},
			Address: req.SessionAddress,
			Name:    req.SessionAddress,
		}

		err = d.userRepo.Create(ctx, user)
		if err != nil {
			return nil, errorx.NewGeneric(err, "Cannot create user")
		}
	}

	token, err := ctx.AccessTokenEngine().Generate(user.ID, model.AccessToken{
		ID:      user.ID,
		Name:    user.Name,
		Address: user.Address,
	})
	if err != nil {
		return nil, errorx.NewGeneric(err, "Cannot generate access token")
	}

	return &model.WalletVerifyResponse{AccessToken: token}, nil
}
