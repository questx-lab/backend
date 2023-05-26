package domain

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
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
	OAuth2Verify(context.Context, *model.OAuth2VerifyRequest) (*model.OAuth2VerifyResponse, error)
	OAuth2IDVerify(context.Context, *model.OAuth2IDVerifyRequest) (*model.OAuth2IDVerifyResponse, error)
	OAuth2Link(context.Context, *model.OAuth2LinkRequest) (*model.OAuth2LinkResponse, error)
	WalletLogin(context.Context, *model.WalletLoginRequest) (*model.WalletLoginResponse, error)
	WalletVerify(context.Context, *model.WalletVerifyRequest) (*model.WalletVerifyResponse, error)
	WalletLink(context.Context, *model.WalletLinkRequest) (*model.WalletLinkResponse, error)
	TelegramLink(context.Context, *model.TelegramLinkRequest) (*model.TelegramLinkResponse, error)
	Refresh(context.Context, *model.RefreshTokenRequest) (*model.RefreshTokenResponse, error)
}

type authDomain struct {
	hasSuperAdmin      bool
	hasSuperAdminMutex sync.Mutex

	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	oauth2Repo       repository.OAuth2Repository
	oauth2Services   []authenticator.IOAuth2Service
}

func NewAuthDomain(
	ctx context.Context,
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	oauth2Repo repository.OAuth2Repository,
	oauth2Cfgs ...config.OAuth2Config,
) AuthDomain {
	oauth2Services := make([]authenticator.IOAuth2Service, len(oauth2Cfgs))
	for i, cfg := range oauth2Cfgs {
		oauth2Services[i] = authenticator.NewOAuth2Service(ctx, cfg)
	}

	return &authDomain{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		oauth2Repo:       oauth2Repo,
		oauth2Services:   oauth2Services,
	}
}

func (d *authDomain) OAuth2Verify(
	ctx context.Context, req *model.OAuth2VerifyRequest,
) (*model.OAuth2VerifyResponse, error) {
	service, ok := d.getOAuth2Service(req.Type)
	if !ok {
		return nil, errorx.New(errorx.BadRequest, "Unsupported type %s", req.Type)
	}

	serviceUserID, err := service.GetUserID(ctx, req.AccessToken)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot verify access token: %v", err)
		return nil, errorx.Unknown
	}

	user, accessToken, refreshToken, err := d.generateTokensWithServiceUserID(ctx, service, serviceUserID)
	if err != nil {
		return nil, err
	}

	clientUser := model.User{
		ID:           user.ID,
		Address:      user.Address.String,
		Name:         user.Name,
		Role:         string(user.Role),
		ReferralCode: user.ReferralCode,
		Services:     make(map[string]string),
		IsNewUser:    user.IsNewUser,
		AvatarURL:    user.ProfilePicture,
	}

	oauth2Records, err := d.oauth2Repo.GetAllByUserID(ctx, user.ID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get all service user ids: %v", err)
		return nil, errorx.Unknown
	}

	for _, record := range oauth2Records {
		clientUser.Services[record.Service] = record.ServiceUserID
	}

	return &model.OAuth2VerifyResponse{
		User:         clientUser,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (d *authDomain) OAuth2IDVerify(
	ctx context.Context, req *model.OAuth2IDVerifyRequest,
) (*model.OAuth2IDVerifyResponse, error) {
	service, ok := d.getOAuth2Service(req.Type)
	if !ok {
		return nil, errorx.New(errorx.BadRequest, "Unsupported type %s", req.Type)
	}

	serviceUserID, err := service.VerifyIDToken(ctx, req.IDToken)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot verify id token: %v", err)
		return nil, errorx.Unknown
	}

	user, accessToken, refreshToken, err := d.generateTokensWithServiceUserID(ctx, service, serviceUserID)
	if err != nil {
		return nil, err
	}

	clientUser := model.User{
		ID:           user.ID,
		Address:      user.Address.String,
		Name:         user.Name,
		Role:         string(user.Role),
		ReferralCode: user.ReferralCode,
		Services:     make(map[string]string),
		IsNewUser:    user.IsNewUser,
		AvatarURL:    user.ProfilePicture,
	}

	oauth2Records, err := d.oauth2Repo.GetAllByUserID(ctx, user.ID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get all service user ids: %v", err)
		return nil, errorx.Unknown
	}

	for _, record := range oauth2Records {
		clientUser.Services[record.Service] = record.ServiceUserID
	}

	return &model.OAuth2IDVerifyResponse{
		User:         clientUser,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (d *authDomain) OAuth2Link(
	ctx context.Context, req *model.OAuth2LinkRequest,
) (*model.OAuth2LinkResponse, error) {
	service, ok := d.getOAuth2Service(req.Type)
	if !ok {
		return nil, errorx.New(errorx.BadRequest, "Unsupported type %s", req.Type)
	}

	serviceUserID, err := service.GetUserID(ctx, req.AccessToken)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot verify access token: %v", err)
		return nil, errorx.Unknown
	}

	_, err = d.userRepo.GetByServiceUserID(ctx, service.Service(), serviceUserID)
	if err == nil {
		return nil, errorx.New(errorx.AlreadyExists, "This %s account has been linked before", service.Service())
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		xcontext.Logger(ctx).Errorf("Cannot get service user id: %v", err)
		return nil, errorx.Unknown
	}

	err = d.oauth2Repo.Create(ctx, &entity.OAuth2{
		UserID:        xcontext.RequestUserID(ctx),
		Service:       service.Service(),
		ServiceUserID: serviceUserID,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot link user with %s: %v", service.Service(), err)
		return nil, errorx.Unknown
	}

	return &model.OAuth2LinkResponse{}, nil
}

func (d *authDomain) WalletLogin(
	ctx context.Context, req *model.WalletLoginRequest,
) (*model.WalletLoginResponse, error) {
	nonce, err := crypto.GenerateRandomString()
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot generate random string: %v", err)
		return nil, errorx.Unknown
	}

	return &model.WalletLoginResponse{Address: req.Address, Nonce: nonce}, nil
}

func (d *authDomain) WalletVerify(
	ctx context.Context, req *model.WalletVerifyRequest,
) (*model.WalletVerifyResponse, error) {
	if err := d.verifyWalletAnswer(ctx, req.Signature, req.SessionNonce, req.SessionAddress); err != nil {
		return nil, err
	}

	user, err := d.userRepo.GetByAddress(ctx, req.SessionAddress)
	if err != nil {
		user = &entity.User{
			Base:    entity.Base{ID: uuid.NewString()},
			Address: sql.NullString{Valid: true, String: req.SessionAddress},
			Name:    req.SessionAddress,
		}

		err = d.createUser(ctx, user)
		if err != nil {
			return nil, err
		}
	}

	refreshToken, err := d.generateRefreshToken(ctx, user.ID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot generate refresh token: %v", err)
		return nil, errorx.Unknown
	}

	accessToken, err := xcontext.TokenEngine(ctx).Generate(
		xcontext.Configs(ctx).Auth.AccessToken.Expiration,
		model.AccessToken{
			ID:      user.ID,
			Name:    user.Name,
			Address: user.Address.String,
		})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot generate access token: %v", err)
		return nil, errorx.Unknown
	}

	clientUser := model.User{
		ID:           user.ID,
		Address:      user.Address.String,
		Name:         user.Name,
		Role:         string(user.Role),
		ReferralCode: user.ReferralCode,
		Services:     make(map[string]string),
		IsNewUser:    user.IsNewUser,
		AvatarURL:    user.ProfilePicture,
	}

	oauth2Records, err := d.oauth2Repo.GetAllByUserID(ctx, user.ID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get all service user ids: %v", err)
		return nil, errorx.Unknown
	}

	for _, record := range oauth2Records {
		clientUser.Services[record.Service] = record.ServiceUserID
	}

	return &model.WalletVerifyResponse{
		User:         clientUser,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (d *authDomain) WalletLink(
	ctx context.Context, req *model.WalletLinkRequest,
) (*model.WalletLinkResponse, error) {
	if err := d.verifyWalletAnswer(ctx, req.Signature, req.SessionNonce, req.SessionAddress); err != nil {
		return nil, err
	}

	_, err := d.userRepo.GetByAddress(ctx, req.SessionAddress)
	if err == nil {
		return nil, errorx.New(errorx.AlreadyExists, "This wallet address has been linked before")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		xcontext.Logger(ctx).Errorf("Cannot get service user id: %v", err)
		return nil, errorx.Unknown
	}

	err = d.userRepo.UpdateByID(ctx, xcontext.RequestUserID(ctx), &entity.User{
		Address: sql.NullString{Valid: true, String: req.SessionAddress},
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot link user with address: %v", err)
		return nil, errorx.Unknown
	}

	return &model.WalletLinkResponse{}, nil
}

func (d *authDomain) TelegramLink(
	ctx context.Context, req *model.TelegramLinkRequest,
) (*model.TelegramLinkResponse, error) {
	telegramCfg := xcontext.Configs(ctx).Auth.Telegram
	serviceUserID := telegramCfg.Name + "_" + req.ID
	_, err := d.userRepo.GetByServiceUserID(ctx, telegramCfg.Name, serviceUserID)
	if err == nil {
		return nil, errorx.New(errorx.AlreadyExists, "This telegram account has been linked before")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		xcontext.Logger(ctx).Errorf("Cannot get service user id: %v", err)
		return nil, errorx.Unknown
	}

	authDate := time.Unix(int64(req.AuthDate), 0)
	if time.Since(authDate) > telegramCfg.LoginExpiration {
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
	hashToken := sha256.Sum256([]byte(telegramCfg.BotToken))
	calculatedHMAC := crypto.HMAC(sha256.New, data, hashToken[:])

	if calculatedHMAC != req.Hash {
		return nil, errorx.New(errorx.Unavailable, "Got an invalid hash")
	}

	err = d.oauth2Repo.Create(ctx, &entity.OAuth2{
		UserID:        xcontext.RequestUserID(ctx),
		Service:       telegramCfg.Name,
		ServiceUserID: serviceUserID,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot link user with telegram: %v", err)
		return nil, errorx.Unknown
	}

	return &model.TelegramLinkResponse{}, nil
}

func (d *authDomain) Refresh(
	ctx context.Context, req *model.RefreshTokenRequest,
) (*model.RefreshTokenResponse, error) {
	// Verify the refresh token from client.
	refreshToken := model.RefreshToken{}
	err := xcontext.TokenEngine(ctx).Verify(req.RefreshToken, &refreshToken)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Failed to verify refresh token: %v", err)
		return nil, errorx.Unknown
	}

	// Load the storage refresh token from database.
	hashedFamily := crypto.SHA256([]byte(refreshToken.Family))
	storageToken, err := d.refreshTokenRepo.Get(ctx, hashedFamily)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get refresh token family %s: %v", refreshToken.Family, err)
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
			xcontext.Logger(ctx).Errorf("Cannot delete refresh token: %v", err)
			return nil, errorx.Unknown
		}

		return nil, errorx.New(errorx.StolenDectected,
			"Your refresh token will be revoked because it is detected as stolen")
	}

	// Rotate the refresh token by increasing counter by 1.
	err = d.refreshTokenRepo.Rotate(ctx, hashedFamily)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot rotate the refresh token: %v", err)
		return nil, errorx.Unknown
	}

	// Everything is ok, generate refresh token and access token.
	newRefreshToken, err := xcontext.TokenEngine(ctx).Generate(
		xcontext.Configs(ctx).Auth.RefreshToken.Expiration,
		model.RefreshToken{
			Family:  refreshToken.Family,
			Counter: refreshToken.Counter + 1,
		})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot generate refresh token: %v", err)
		return nil, errorx.Unknown
	}

	user, err := d.userRepo.GetByID(ctx, storageToken.UserID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get user: %v", err)
		return nil, errorx.Unknown
	}

	newAccessToken, err := xcontext.TokenEngine(ctx).Generate(
		xcontext.Configs(ctx).Auth.AccessToken.Expiration,
		model.AccessToken{
			ID:      user.ID,
			Name:    user.Name,
			Address: user.Address.String,
		})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot generate access token: %v", err)
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

func (d *authDomain) generateRefreshToken(ctx context.Context, userID string) (string, error) {
	refreshTokenFamily, err := crypto.GenerateRandomString()
	if err != nil {
		return "", err
	}

	refreshToken, err := xcontext.TokenEngine(ctx).Generate(
		xcontext.Configs(ctx).Auth.RefreshToken.Expiration,
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
		Expiration: time.Now().Add(xcontext.Configs(ctx).Auth.RefreshToken.Expiration),
	})
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func (d *authDomain) verifyWalletAnswer(ctx context.Context, hexSignature, sessionNonce, sessionAddress string) error {
	hash := accounts.TextHash([]byte(sessionNonce))
	signature, err := hexutil.Decode(hexSignature)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot decode signature: %v", err)
		return errorx.Unknown
	}

	if signature[ethcrypto.RecoveryIDOffset] == 27 || signature[ethcrypto.RecoveryIDOffset] == 28 {
		signature[ethcrypto.RecoveryIDOffset] -= 27 // Transform yellow paper V from 27/28 to 0/1
	}

	recovered, err := ethcrypto.SigToPub(hash, signature)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot recover signature to address: %v", err)
		return errorx.Unknown
	}

	recoveredAddr := ethcrypto.PubkeyToAddress(*recovered)
	if !bytes.Equal(recoveredAddr.Bytes(), ethcommon.HexToAddress(sessionAddress).Bytes()) {
		return errorx.New(errorx.BadRequest, "Mismatched address")
	}

	return nil
}

func (d *authDomain) generateTokensWithServiceUserID(
	ctx context.Context, service authenticator.IOAuth2Service, serviceUserID string,
) (*entity.User, string, string, error) {
	user, err := d.userRepo.GetByServiceUserID(ctx, service.Service(), serviceUserID)
	if err != nil {
		ctx = xcontext.WithDBTransaction(ctx)
		defer xcontext.WithRollbackDBTransaction(ctx)

		user = &entity.User{
			Base:    entity.Base{ID: uuid.NewString()},
			Address: sql.NullString{Valid: false},
			Name:    serviceUserID,
		}

		err = d.createUser(ctx, user)
		if err != nil {
			return nil, "", "", err
		}

		err = d.oauth2Repo.Create(ctx, &entity.OAuth2{
			UserID:        user.ID,
			Service:       service.Service(),
			ServiceUserID: serviceUserID,
		})
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot register user with service: %v", err)
			return nil, "", "", errorx.New(errorx.AlreadyExists,
				"This %s account was already registered with another user", service.Service())
		}

		ctx = xcontext.WithCommitDBTransaction(ctx)
	}

	refreshToken, err := d.generateRefreshToken(ctx, user.ID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot generate refresh token: %v", err)
		return nil, "", "", errorx.Unknown
	}

	accessToken, err := xcontext.TokenEngine(ctx).Generate(
		xcontext.Configs(ctx).Auth.AccessToken.Expiration,
		model.AccessToken{
			ID:      user.ID,
			Name:    user.Name,
			Address: user.Address.String,
		})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot generate access token: %v", err)
		return nil, "", "", errorx.Unknown
	}

	return user, accessToken, refreshToken, nil
}

func (d *authDomain) createUser(ctx context.Context, user *entity.User) error {
	user.Role = entity.RoleUser
	user.IsNewUser = true
	user.ReferralCode = crypto.GenerateRandomAlphabet(9)

	if !d.hasSuperAdmin {
		d.hasSuperAdminMutex.Lock()
		defer d.hasSuperAdminMutex.Unlock()

		if !d.hasSuperAdmin {
			count, err := d.userRepo.Count(ctx)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot count number of user records: %v", err)
				return errorx.Unknown
			}

			if count == 0 {
				user.Role = entity.RoleSuperAdmin
			}
		}
	}

	if err := d.userRepo.Create(ctx, user); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create user: %v", err)
		return errorx.Unknown
	}

	if !d.hasSuperAdmin {
		d.hasSuperAdmin = true
	}

	return nil
}
