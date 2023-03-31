package domain

import (
	"context"
	"testing"

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

func Test_oauth2Domain_Callback_DuplicateServiceID(t *testing.T) {
	// Mock oauth2 returns a specific service user id.
	duplicated_id := "duplicated_service_user_id"
	oauth2Config := testutil.NewMockOAuth2()
	oauth2Config.VerifyIDTokenFunc = func(ctx context.Context, token *oauth2.Token) (string, error) {
		return duplicated_id, nil
	}

	// Generate database environment. DO NOT create fixture db here.
	ctx := testutil.NewMockContext()
	userRepo := repository.NewUserRepository()
	oauth2Repo := repository.NewOAuth2Repository()

	// Create a new oauth2 domain.
	domain := oauth2Domain{
		userRepo:      userRepo,
		oauth2Repo:    oauth2Repo,
		oauth2Configs: []authenticator.IOAuth2Config{oauth2Config},
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
	_, err = domain.Callback(ctx, &model.OAuth2CallbackRequest{Type: oauth2Config.Name})
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
