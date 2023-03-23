package domain

import (
	"testing"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"

	"github.com/stretchr/testify/assert"
)

func Test_projectDomain_Create(t *testing.T) {
	Initialized()
	req := &model.CreateProjectRequest{
		Name:     "test",
		Twitter:  "https://twitter.com/hashtag/Breaking2",
		Discord:  "https://discord.com/hashtag/Breaking2",
		Telegram: "https://telegram.com/",
	}
	ctx := NewMockContext()
	resp, err := projectdomain.Create(ctx, req)
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
