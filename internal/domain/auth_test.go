package domain

import (
	"context"
	"testing"
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/authenticator"
	"github.com/questx-lab/backend/pkg/crypto"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func Test_authDomain_OAuth2Verify_DuplicateServiceID(t *testing.T) {
	// Mock oauth2 returns a specific service user id.
	duplicated_id := "duplicated_service_user_id"
	oauth2Config := testutil.NewMockOAuth2("example")
	oauth2Config.GetUserIDFunc = func(context.Context, string) (authenticator.OAuth2User, error) {
		return authenticator.OAuth2User{ID: duplicated_id}, nil
	}

	// Generate database environment. DO NOT create fixture db here.
	ctx := testutil.MockContext(t)
	userRepo := repository.NewUserRepository(testutil.RedisClient(ctx))
	oauth2Repo := repository.NewOAuth2Repository()
	refreshTokenRepo := repository.NewRefreshTokenRepository()

	// Create a new oauth2 domain.
	domain := authDomain{
		userRepo:         userRepo,
		oauth2Repo:       oauth2Repo,
		oauth2Services:   []authenticator.IOAuth2Service{oauth2Config},
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
	_, err = domain.OAuth2Verify(ctx, &model.OAuth2VerifyRequest{
		Type:        oauth2Config.Name,
		AccessToken: "foo",
	})
	var errx errorx.Error
	require.ErrorAs(t, err, &errx)
	require.Equal(t, errorx.AlreadyExists, errx.Code)

	// The user record is inserted before oauth2 record. But the transaction ensures that user
	// record will be rollbacked after the error.
	// So there is no record in user table because this is a fresh db.
	var user entity.User
	err = xcontext.DB(ctx).First(&user).Error
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func Test_authDomain_Refresh(t *testing.T) {
	ctx := testutil.MockContext(t)
	testutil.CreateFixtureDb(ctx)

	domain := &authDomain{
		userRepo:         repository.NewUserRepository(testutil.RedisClient(ctx)),
		refreshTokenRepo: repository.NewRefreshTokenRepository(),
	}

	refreshTokenObj := model.RefreshToken{
		Family:  "Foo",
		Counter: 0,
	}

	err := domain.refreshTokenRepo.Create(ctx, &entity.RefreshToken{
		UserID:     testutil.User1.ID,
		Family:     crypto.SHA256([]byte(refreshTokenObj.Family)),
		Counter:    0,
		Expiration: time.Now().Add(time.Minute),
	})
	require.NoError(t, err)

	refreshToken, err := xcontext.TokenEngine(ctx).Generate(time.Minute, refreshTokenObj)
	require.NoError(t, err)

	// Successfully for the first refresh.
	resp, err := domain.Refresh(ctx, &model.RefreshTokenRequest{RefreshToken: refreshToken})
	require.NoError(t, err)

	// Verify access token.
	accessToken := model.AccessToken{}
	err = xcontext.TokenEngine(ctx).Verify(resp.AccessToken, &accessToken)
	require.NoError(t, err)
	require.Equal(t, testutil.User1.ID, accessToken.ID)

	// Detect stolen for the second refresh, the refresh token will be deleted after this call.
	_, err = domain.Refresh(ctx, &model.RefreshTokenRequest{RefreshToken: refreshToken})
	require.Error(t, err)
	require.Equal(t, "Your refresh token will be revoked because it is detected as stolen", err.Error())

	// Not found refresh token for the third refresh.
	_, err = domain.Refresh(ctx, &model.RefreshTokenRequest{RefreshToken: refreshToken})
	require.Error(t, err)
	require.Equal(t, "Request failed", err.Error())
}
