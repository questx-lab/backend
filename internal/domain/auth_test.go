package domain

import (
	"context"
	"testing"
	"time"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/authenticator"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

func Test_authDomain_OAuth2Callback_DuplicateServiceID(t *testing.T) {
	// Mock oauth2 returns a specific service user id.
	duplicated_id := "duplicated_service_user_id"
	oauth2Config := testutil.NewMockOAuth2("example")
	oauth2Config.VerifyIDTokenFunc = func(ctx context.Context, token *oauth2.Token) (string, error) {
		return duplicated_id, nil
	}

	// Generate database environment. DO NOT create fixture db here.
	ctx := testutil.NewMockContext()
	userRepo := repository.NewUserRepository()
	oauth2Repo := repository.NewOAuth2Repository()
	refreshTokenRepo := repository.NewRefreshTokenRepository()

	// Create a new oauth2 domain.
	domain := authDomain{
		userRepo:         userRepo,
		oauth2Repo:       oauth2Repo,
		oauth2Configs:    []authenticator.IOAuth2Config{oauth2Config},
		refreshTokenRepo: refreshTokenRepo,
	}

	// Insert a record with the service user id is duplicated with the one returned by oauth2
	// service.
	err := oauth2Repo.Create(ctx, &entity.OAuth2{
		UserID:        "user-id",
		Service:       oauth2Config.Name,
		ServiceUserID: duplicated_id,
	})
	require.NoError(t, err)

	// The callback method cannot process this request because it failed to insert a record with a
	// duplicated field in oauth2 table.
	_, err = domain.OAuth2Callback(ctx, &model.OAuth2CallbackRequest{Type: oauth2Config.Name})
	var errx errorx.Error
	require.ErrorAs(t, err, &errx)
	require.Equal(t, errorx.AlreadyExists, errx.Code)

	// The user record is inserted before oauth2 record. But the transaction ensures that user
	// record will be rollbacked after the error.
	// So there is no record in user table because this is a fresh db.
	var user entity.User
	err = ctx.DB().First(&user).Error
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func Test_authDomain_Refresh(t *testing.T) {
	ctx := testutil.NewMockContext()
	testutil.CreateFixtureDb(ctx)

	domain := &authDomain{
		refreshTokenRepo: repository.NewRefreshTokenRepository(),
	}

	refreshTokenObj := model.RefreshToken{
		Family:  "Foo",
		Counter: 0,
	}

	err := domain.refreshTokenRepo.Create(ctx, &entity.RefreshToken{
		UserID:     testutil.User1.ID,
		Family:     common.Hash([]byte(refreshTokenObj.Family)),
		Counter:    0,
		Expiration: time.Now().Add(time.Minute),
	})
	require.NoError(t, err)

	refreshToken, err := ctx.TokenEngine().Generate(time.Minute, refreshTokenObj)
	require.NoError(t, err)

	_, err = domain.Refresh(ctx, &model.RefreshTokenRequest{RefreshToken: refreshToken})
	require.NoError(t, err)
}
