package domain

import (
	"bytes"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/authenticator"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type AuthDomain interface {
	OAuth2Login(xcontext.Context, *model.OAuth2LoginRequest) (*model.OAuth2LoginResponse, error)
	OAuth2Callback(xcontext.Context, *model.OAuth2CallbackRequest) (*model.OAuth2CallbackResponse, error)
	WalletLogin(xcontext.Context, *model.WalletLoginRequest) (*model.WalletLoginResponse, error)
	WalletVerify(xcontext.Context, *model.WalletVerifyRequest) (*model.WalletVerifyResponse, error)
	Refresh(xcontext.Context, *model.RefreshTokenRequest) (*model.RefreshTokenResponse, error)
}

type authDomain struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	oauth2Repo       repository.OAuth2Repository
	oauth2Configs    []authenticator.IOAuth2Config
}

func NewAuthDomain(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	oauth2Repo repository.OAuth2Repository,
	authenticators []authenticator.IOAuth2Config,
) AuthDomain {
	return &authDomain{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		oauth2Repo:       oauth2Repo,
		oauth2Configs:    authenticators,
	}
}

func (d *authDomain) OAuth2Login(
	ctx xcontext.Context, req *model.OAuth2LoginRequest,
) (*model.OAuth2LoginResponse, error) {
	authenticator, ok := d.getAuthenticator(req.Type)
	if !ok {
		return nil, errorx.New(errorx.BadRequest, "Unsupported type %s", req.Type)
	}

	state, err := common.GenerateRandomString()
	if err != nil {
		ctx.Logger().Errorf("Cannot generate random string: %s", err)
		return nil, errorx.Unknown
	}

	return &model.OAuth2LoginResponse{
		RedirectURL: authenticator.AuthCodeURL(state),
		State:       state,
	}, nil
}

func (d *authDomain) OAuth2Callback(
	ctx xcontext.Context, req *model.OAuth2CallbackRequest,
) (*model.OAuth2CallbackResponse, error) {
	auth, ok := d.getAuthenticator(req.Type)
	if !ok {
		return nil, errorx.New(errorx.BadRequest, "Unsupported type %s", req.Type)
	}

	if req.State != req.SessionState {
		return nil, errorx.New(errorx.BadRequest, "Mismatched state parameter")
	}

	// Exchange an authorization code for a serviceToken.
	serviceToken, err := auth.Exchange(ctx, req.Code)
	if err != nil {
		ctx.Logger().Warnf("Cannot exchange authorization code: %v", err)
		return nil, errorx.Unknown
	}

	serviceUserID, err := auth.VerifyIDToken(ctx, serviceToken)
	if err != nil {
		ctx.Logger().Warnf("Cannot verify id token: %v", err)
		return nil, errorx.Unknown
	}

	user, err := d.userRepo.GetByServiceUserID(ctx, auth.Service(), serviceUserID)
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
			Service:       auth.Service(),
			ServiceUserID: serviceUserID,
		})
		if err != nil {
			ctx.Logger().Errorf("Cannot register user with service: %v", err)
			return nil, errorx.New(errorx.AlreadyExists,
				"This %s account was already registered with another user", auth.Service())
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

	return &model.OAuth2CallbackResponse{
		RedirectURL:  ctx.Configs().Auth.CallbackURL,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (d *authDomain) WalletLogin(
	ctx xcontext.Context, req *model.WalletLoginRequest,
) (*model.WalletLoginResponse, error) {
	nonce, err := common.GenerateRandomString()
	if err != nil {
		ctx.Logger().Errorf("Cannot generate random string: %v", err)
		return nil, errorx.Unknown
	}

	return &model.WalletLoginResponse{Address: req.Address, Nonce: nonce}, nil
}

func (d *authDomain) WalletVerify(
	ctx xcontext.Context, req *model.WalletVerifyRequest,
) (*model.WalletVerifyResponse, error) {
	hash := accounts.TextHash([]byte(req.SessionNonce))
	signature, err := hexutil.Decode(req.Signature)
	if err != nil {
		ctx.Logger().Errorf("Cannot decode signature: %v", err)
		return nil, errorx.Unknown
	}

	if signature[crypto.RecoveryIDOffset] == 27 || signature[crypto.RecoveryIDOffset] == 28 {
		signature[crypto.RecoveryIDOffset] -= 27 // Transform yellow paper V from 27/28 to 0/1
	}

	recovered, err := crypto.SigToPub(hash, signature)
	if err != nil {
		ctx.Logger().Errorf("Cannot recover signature to address: %v", err)
		return nil, errorx.Unknown
	}

	recoveredAddr := crypto.PubkeyToAddress(*recovered)
	if !bytes.Equal(recoveredAddr.Bytes(), ethcommon.HexToAddress(req.SessionAddress).Bytes()) {
		return nil, errorx.New(errorx.BadRequest, "Mismatched address")
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
	storageToken, err := d.refreshTokenRepo.Get(ctx, common.Hash([]byte(refreshToken.Family)))
	if err != nil {
		ctx.Logger().Errorf("Cannot get refresh token family %s: %v", refreshToken.Family, err)
		return nil, errorx.Unknown
	}

	// Check the expiration of storage refresh token.
	if storageToken.Expiration.Before(time.Now()) {
		return nil, errorx.New(errorx.TokenExpired, "Your refresh token is expired")
	}

	// Check if refresh token is stolen or invalid.
	if refreshToken.Counter != storageToken.Counter {
		err = d.refreshTokenRepo.Delete(ctx, refreshToken.Family)
		if err != nil {
			ctx.Logger().Errorf("Cannot delete refresh token: %v", err)
			return nil, errorx.Unknown
		}

		return nil, errorx.New(errorx.StolenDectected,
			"Your refresh token will be revoked because it is detected as stolen")
	}

	// Rotate the refresh token by increasing index by 1.
	err = d.refreshTokenRepo.Rotate(ctx, common.Hash([]byte(refreshToken.Family)))
	if err != nil {
		ctx.Logger().Errorf("Cannot rotate the refresh token: %v", err)
		return nil, errorx.Unknown
	}

	// Everything is ok, generate refresh token and access token.
	newRefreshToken, err := ctx.TokenEngine().Generate(
		ctx.Configs().Auth.RefreshToken.Expiration,
		model.RefreshToken{
			Family:  storageToken.Family,
			Counter: storageToken.Counter + 1,
		})
	if err != nil {
		ctx.Logger().Errorf("Cannot generate refresh token: %v", err)
		return nil, errorx.Unknown
	}

	newAccessToken, err := ctx.TokenEngine().Generate(
		ctx.Configs().Auth.AccessToken.Expiration,
		model.AccessToken{
			ID:      storageToken.User.ID,
			Name:    storageToken.User.Name,
			Address: storageToken.User.Address,
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

func (d *authDomain) getAuthenticator(service string) (authenticator.IOAuth2Config, bool) {
	for i := range d.oauth2Configs {
		if d.oauth2Configs[i].Service() == service {
			return d.oauth2Configs[i], true
		}
	}
	return &authenticator.OAuth2Config{}, false
}

func (d *authDomain) generateRefreshToken(ctx xcontext.Context, userID string) (string, error) {
	refreshTokenFamily, err := common.GenerateRandomString()
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
		Family:     common.Hash([]byte(refreshTokenFamily)),
		Counter:    0,
		Expiration: time.Now().Add(ctx.Configs().Auth.RefreshToken.Expiration),
	})
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}
