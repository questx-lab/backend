package domain

import (
	"testing"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/testutil"

	"github.com/stretchr/testify/require"
)

func Test_projectDomain_Create(t *testing.T) {
	db := testutil.GetEmptyTestDb()
	projectRepo := repository.NewProjectRepository(db)
	domain := NewProjectDomain(projectRepo)
	validUserID := "valid-user-id"
	req := &model.CreateProjectRequest{
		Name:     "test",
		Twitter:  "https://twitter.com/hashtag/Breaking2",
		Discord:  "https://discord.com/hashtag/Breaking2",
		Telegram: "https://telegram.com/",
	}
	ctx := testutil.NewMockContextWithUserID(validUserID)
	resp, err := domain.Create(ctx, req)
	require.NoError(t, err)
	require.True(t, resp.Success)

	var result entity.Project
	tx := db.Model(&entity.Project{}).Take(&result, "id", resp.ID)
	require.NoError(t, tx.Error)
	require.Equal(t, result.Name, req.Name)
	require.Equal(t, result.Discord, req.Discord)
	require.Equal(t, result.Twitter, req.Twitter)
	require.Equal(t, result.Telegram, req.Telegram)
	require.Equal(t, result.CreatedBy, validUserID)
}
