package domain

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/authenticator"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func Test_projectDomain_Create(t *testing.T) {
	Initialized()
	expiration := time.Minute
	req := &model.CreateProjectRequest{
		Name:     "test",
		Twitter:  "https://twitter.com/hashtag/Breaking2",
		Discord:  "https://discord.com/hashtag/Breaking2",
		Telegram: "https://telegram.com/",
	}
	ctx := router.DefaultContext()
	r := httptest.NewRequest(http.MethodGet, "/createProject", nil)
	authenticator := authenticator.NewTokenEngine[model.AccessToken](config.TokenConfigs{
		Secret:     "secret",
		Expiration: expiration,
	})
	userID := "valid-user-id"
	tkn, err := authenticator.Generate(userID, model.AccessToken{
		ID: userID,
	})
	assert.NoError(t, err)
	r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tkn))
	ctx.SetRequest(r)
	ctx.SetAccessTokenEngine(authenticator)
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
	assert.Equal(t, result.CreatedBy, userID)
}
