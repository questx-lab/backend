package domain

import (
	"context"
	"testing"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	db := testutil.GetTestDb(t)

	userId := "user1"
	accessToken := model.AccessToken{
		ID:      userId,
		Name:    "google",
		Address: "address",
	}

	repo := repository.NewUserRepository(db)
	repo.Create(context.Background(), &entity.User{
		ID:      userId,
		Address: "address",
	})

	_, err := repo.RetrieveByID(context.Background(), "user1")
	require.Nil(t, err)

	userDomain := NewUserDomain(repository.NewUserRepository(db))
	res, err := userDomain.GetUser(
		*testutil.GetUserContextWithAccessToken(t, accessToken),
		&model.GetUserRequest{
			ID: userId,
		},
	)
	require.Nil(t, err)

	require.Equal(t, &model.GetUserResponse{
		ID:      userId,
		Address: "address",
	}, res)
}
