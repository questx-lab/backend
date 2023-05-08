package domain

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/authenticator"
	"github.com/questx-lab/backend/pkg/crypto"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type AuthDomain interface {
	OAuth2Verify(xcontext.Context, *model.OAuth2VerifyRequest) (*model.OAuth2VerifyResponse, error)
	OAuth2Link(xcontext.Context, *model.OAuth2LinkRequest) (*model.OAuth2LinkResponse, error)
	WalletLogin(xcontext.Context, *model.WalletLoginRequest) (*model.WalletLoginResponse, error)
	WalletVerify(xcontext.Context, *model.WalletVerifyRequest) (*model.WalletVerifyResponse, error)
	WalletLink(xcontext.Context, *model.WalletLinkRequest) (*model.WalletLinkResponse, error)
	TelegramLink(xcontext.Context, *model.TelegramLinkRequest) (*model.TelegramLinkResponse, error)
	Refresh(xcontext.Context, *model.RefreshTokenRequest) (*model.RefreshTokenResponse, error)
}

type authDomain struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	oauth2Repo       repository.OAuth2Repository
	oauth2Services   []authenticator.IOAuth2Service
}

func NewAuthDomain(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	oauth2Repo repository.OAuth2Repository,
	oauth2Cfgs ...config.OAuth2Config,
) AuthDomain {
	oauth2Services := make([]authenticator.IOAuth2Service, len(oauth2Cfgs))
	for i, cfg := range oauth2Cfgs {
		oauth2Services[i] = authenticator.NewOAuth2Service(cfg)
	}

	return &authDomain{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		oauth2Repo:       oauth2Repo,
		oauth2Services:   oauth2Services,
	}
}

func (d *authDomain) OAuth2Verify(
	ctx xcontext.Context, req *model.OAuth2VerifyRequest,
) (*model.OAuth2VerifyResponse, error) {
	service, ok := d.getOAuth2Service(req.Type)
	if !ok {
		return nil, errorx.New(errorx.BadRequest, "Unsupported type %s", req.Type)
	}

	serviceUserID, err := service.GetUserID(ctx, req.AccessToken)
	if err != nil {
		ctx.Logger().Errorf("Cannot verify access token: %v", err)
		return nil, errorx.Unknown
	}

	user, err := d.userRepo.GetByServiceUserID(ctx, service.Service(), serviceUserID)
	if err != nil {
		ctx.BeginTx()
		defer ctx.RollbackTx()

		user = &entity.User{
			Base:    entity.Base{ID: uuid.NewString()},
			Address: "",
			Name:    serviceUserID,
		}

		err = d.userRepo.Create(ctx, user)
		if err != nil {
			ctx.Logger().Errorf("Cannot create user: %v", err)
			return nil, errorx.Unknown
		}

		err = d.oauth2Repo.Create(ctx, &entity.OAuth2{
			UserID:        user.ID,
			Service:       service.Service(),
			ServiceUserID: serviceUserID,
		})
		if err != nil {
			ctx.Logger().Errorf("Cannot register user with service: %v", err)
			return nil, errorx.New(errorx.AlreadyExists,
				"This %s account was already registered with another user", service.Service())
		}

		ctx.CommitTx()
	}

	refreshToken, err := d.generateRefreshToken(ctx, user.ID)
	if err != nil {
		ctx.Logger().Errorf("Cannot generate refresh token: %v", err)
		return nil, errorx.Unknown
	}

	accessToken, err := ctx.TokenEngine().Generate(
		ctx.Configs().Auth.AccessToken.Expiration,
		model.AccessToken{
			ID:      user.ID,
			Name:    user.Name,
			Address: user.Address,
		})
	if err != nil {
		ctx.Logger().Errorf("Cannot generate access token: %v", err)
		return nil, errorx.Unknown
	}

	return &model.OAuth2VerifyResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (d *authDomain) OAuth2Link(
	ctx xcontext.Context, req *model.OAuth2LinkRequest,
) (*model.OAuth2LinkResponse, error) {
	service, ok := d.getOAuth2Service(req.Type)
	if !ok {
		return nil, errorx.New(errorx.BadRequest, "Unsupported type %s", req.Type)
	}

	serviceUserID, err := service.GetUserID(ctx, req.AccessToken)
	if err != nil {
		ctx.Logger().Errorf("Cannot verify access token: %v", err)
		return nil, errorx.Unknown
	}

	_, err = d.userRepo.GetByServiceUserID(ctx, service.Service(), serviceUserID)
	if err == nil {
		return nil, errorx.New(errorx.AlreadyExists, "This %s account has been linked before", service.Service())
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.Logger().Errorf("Cannot get service user id: %v", err)
		return nil, errorx.Unknown
	}

	err = d.oauth2Repo.Create(ctx, &entity.OAuth2{
		UserID:        xcontext.GetRequestUserID(ctx),
		Service:       ctx.Configs().Auth.Telegram.Name,
		ServiceUserID: serviceUserID,
	})
	if err != nil {
		ctx.Logger().Errorf("Cannot link user with %s: %v", service.Service(), err)
		return nil, errorx.Unknown
	}

	return &model.OAuth2LinkResponse{}, nil
}

func (d *authDomain) WalletLogin(
	ctx xcontext.Context, req *model.WalletLoginRequest,
) (*model.WalletLoginResponse, error) {
	nonce, err := crypto.GenerateRandomString()
	if err != nil {
		ctx.Logger().Errorf("Cannot generate random string: %v", err)
		return nil, errorx.Unknown
	}

	return &model.WalletLoginResponse{Address: req.Address, Nonce: nonce}, nil
}

func (d *authDomain) WalletVerify(
	ctx xcontext.Context, req *model.WalletVerifyRequest,
) (*model.WalletVerifyResponse, error) {
	if err := d.verifyWalletAnswer(ctx, req.Signature, req.SessionNonce, req.SessionAddress); err != nil {
		return nil, err
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
			ctx.Logger().Errorf("Cannot create user: %v", err)
			return nil, errorx.Unknown
		}
	}

	refreshToken, err := d.generateRefreshToken(ctx, user.ID)
	if err != nil {
		ctx.Logger().Errorf("Cannot generate refresh token: %v", err)
		return nil, errorx.Unknown
	}

	accessToken, err := ctx.TokenEngine().Generate(
		ctx.Configs().Auth.AccessToken.Expiration,
		model.AccessToken{
			ID:      user.ID,
			Name:    user.Name,
			Address: user.Address,
		})
	if err != nil {
		ctx.Logger().Errorf("Cannot generate access token: %v", err)
		return nil, errorx.Unknown
	}

	return &model.WalletVerifyResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (d *authDomain) WalletLink(
	ctx xcontext.Context, req *model.WalletLinkRequest,
) (*model.WalletLinkResponse, error) {
	if err := d.verifyWalletAnswer(ctx, req.Signature, req.SessionNonce, req.SessionAddress); err != nil {
		return nil, err
	}

	_, err := d.userRepo.GetByAddress(ctx, req.SessionAddress)
	if err == nil {
		return nil, errorx.New(errorx.AlreadyExists, "This wallet address has been linked before")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.Logger().Errorf("Cannot get service user id: %v", err)
		return nil, errorx.Unknown
	}

	err = d.userRepo.UpdateByID(ctx, xcontext.GetRequestUserID(ctx), &entity.User{
		Address: req.SessionAddress,
	})
	if err != nil {
		ctx.Logger().Errorf("Cannot link user with address: %v", err)
		return nil, errorx.Unknown
	}

	return &model.WalletLinkResponse{}, nil
}

func (d *authDomain) TelegramLink(
	ctx xcontext.Context, req *model.TelegramLinkRequest,
) (*model.TelegramLinkResponse, error) {
	serviceUserID := ctx.Configs().Auth.Telegram.Name + "_" + req.ID
	_, err := d.userRepo.GetByServiceUserID(ctx, ctx.Configs().Auth.Telegram.Name, serviceUserID)
	if err == nil {
		return nil, errorx.New(errorx.AlreadyExists, "This telegram account has been linked before")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.Logger().Errorf("Cannot get service user id: %v", err)
		return nil, errorx.Unknown
	}

	authDate := time.Unix(int64(req.AuthDate), 0)
	if time.Since(authDate) > ctx.Configs().Auth.Telegram.LoginExpiration {
		return nil, errorx.New(errorx.BadRequest, "The authentication information is expired")
	}

	fields := []string{}
	fields = append(fields, fmt.Sprintf("auth_date=%d", req.AuthDate))
	fields = append(fields, fmt.Sprintf("first_name=%s", req.FirstName))
	fields = append(fields, fmt.Sprintf("id=%s", req.ID))
	fields = append(fields, fmt.Sprintf("last_name=%s", req.LastName))
	fields = append(fields, fmt.Sprintf("photo_url=%s", req.PhotoURL))
	fields = append(fields, fmt.Sprintf("username=%s", req.Username))
	data := []byte(strings.Join(fields, "\n"))
	hashToken := sha256.Sum256([]byte(ctx.Configs().Auth.Telegram.BotToken))
	calculatedHMAC := crypto.HMAC(sha256.New, data, hashToken[:])

	if calculatedHMAC != req.Hash {
		return nil, errorx.New(errorx.Unavailable, "Got an invalid hash")
	}

	err = d.oauth2Repo.Create(ctx, &entity.OAuth2{
		UserID:        xcontext.GetRequestUserID(ctx),
		Service:       ctx.Configs().Auth.Telegram.Name,
		ServiceUserID: serviceUserID,
	})
	if err != nil {
		ctx.Logger().Errorf("Cannot link user with telegram: %v", err)
		return nil, errorx.Unknown
	}

	return &model.TelegramLinkResponse{}, nil
}

func (d *authDomain) Refresh(
	ctx xcontext.Context, req *model.RefreshTokenRequest,
) (*model.RefreshTokenResponse, error) {
	// Verify the refresh token from client.
	var refreshToken model.RefreshToken
	err := ctx.TokenEngine().Verify(req.RefreshToken, &refreshToken)
	if err != nil {
		ctx.Logger().Debugf("Failed to verify refresh token: %v", err)
		return nil, errorx.Unknown
	}

	// Load the storage refresh token from database.
	hashedFamily := crypto.SHA256([]byte(refreshToken.Family))
	storageToken, err := d.refreshTokenRepo.Get(ctx, hashedFamily)
	if err != nil {
		ctx.Logger().Errorf("Cannot get refresh token family %s: %v", refreshToken.Family, err)
		return nil, errorx.Unknown
	}

	// Check the expiration of storage refresh token.
	if storageToken.Expiration.Before(time.Now()) {
		return nil, errorx.New(errorx.TokenExpired, "Your refresh token is expired")
	}

	// Check if refresh token is stolen or invalid.
	// NOTE: DO NOT create transaction here. The delete and rotate query is independent.
	if refreshToken.Counter != storageToken.Counter {
		err = d.refreshTokenRepo.Delete(ctx, hashedFamily)
		if err != nil {
			ctx.Logger().Errorf("Cannot delete refresh token: %v", err)
			return nil, errorx.Unknown
		}

		return nil, errorx.New(errorx.StolenDectected,
			"Your refresh token will be revoked because it is detected as stolen")
	}

	// Rotate the refresh token by increasing counter by 1.
	err = d.refreshTokenRepo.Rotate(ctx, hashedFamily)
	if err != nil {
		ctx.Logger().Errorf("Cannot rotate the refresh token: %v", err)
		return nil, errorx.Unknown
	}

	// Everything is ok, generate refresh token and access token.
	newRefreshToken, err := ctx.TokenEngine().Generate(
		ctx.Configs().Auth.RefreshToken.Expiration,
		model.RefreshToken{
			Family:  refreshToken.Family,
			Counter: refreshToken.Counter + 1,
		})
	if err != nil {
		ctx.Logger().Errorf("Cannot generate refresh token: %v", err)
		return nil, errorx.Unknown
	}

	user, err := d.userRepo.GetByID(ctx, storageToken.UserID)
	if err != nil {
		ctx.Logger().Errorf("Cannot get user: %v", err)
		return nil, errorx.Unknown
	}

	newAccessToken, err := ctx.TokenEngine().Generate(
		ctx.Configs().Auth.AccessToken.Expiration,
		model.AccessToken{
			ID:      user.ID,
			Name:    user.Name,
			Address: user.Address,
		})
	if err != nil {
		ctx.Logger().Errorf("Cannot generate access token: %v", err)
		return nil, errorx.Unknown
	}

	return &model.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (d *authDomain) getOAuth2Service(service string) (authenticator.IOAuth2Service, bool) {
	for i := range d.oauth2Services {
		if d.oauth2Services[i].Service() == service {
			return d.oauth2Services[i], true
		}
	}
	return nil, false
}

func (d *authDomain) generateRefreshToken(ctx xcontext.Context, userID string) (string, error) {
	refreshTokenFamily, err := crypto.GenerateRandomString()
	if err != nil {
		return "", err
	}

	refreshToken, err := ctx.TokenEngine().Generate(
		ctx.Configs().Auth.RefreshToken.Expiration,
		model.RefreshToken{
			Family:  refreshTokenFamily,
			Counter: 0,
		})
	if err != nil {
		return "", err
	}

	err = d.refreshTokenRepo.Create(ctx, &entity.RefreshToken{
		UserID:     userID,
		Family:     crypto.SHA256([]byte(refreshTokenFamily)),
		Counter:    0,
		Expiration: time.Now().Add(ctx.Configs().Auth.RefreshToken.Expiration),
	})
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func (d *authDomain) verifyWalletAnswer(ctx xcontext.Context, hexSignature, sessionNonce, sessionAddress string) error {
	hash := accounts.TextHash([]byte(sessionNonce))
	signature, err := hexutil.Decode(hexSignature)
	if err != nil {
		ctx.Logger().Errorf("Cannot decode signature: %v", err)
		return errorx.Unknown
	}

	if signature[ethcrypto.RecoveryIDOffset] == 27 || signature[ethcrypto.RecoveryIDOffset] == 28 {
		signature[ethcrypto.RecoveryIDOffset] -= 27 // Transform yellow paper V from 27/28 to 0/1
	}

	recovered, err := ethcrypto.SigToPub(hash, signature)
	if err != nil {
		ctx.Logger().Errorf("Cannot recover signature to address: %v", err)
		return errorx.Unknown
	}

	recoveredAddr := ethcrypto.PubkeyToAddress(*recovered)
	if !bytes.Equal(recoveredAddr.Bytes(), ethcommon.HexToAddress(sessionAddress).Bytes()) {
		return errorx.New(errorx.BadRequest, "Mismatched address")
	}

	return nil
}
