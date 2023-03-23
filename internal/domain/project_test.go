package domain

import (
	"testing"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/testutil"

	"github.com/stretchr/testify/assert"
)

func Test_projectDomain_Create(t *testing.T) {
	db := testutil.GetDatabaseTest()
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
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	var result entity.Project
	tx := db.Model(&entity.Project{}).Take(&result, "id", resp.ID)
	assert.NoError(t, tx.Error)
	assert.Equal(t, result.Name, req.Name)
	assert.Equal(t, result.Discord, req.Discord)
	assert.Equal(t, result.Twitter, req.Twitter)
	assert.Equal(t, result.Telegram, req.Telegram)
	assert.Equal(t, result.CreatedBy, validUserID)
}
