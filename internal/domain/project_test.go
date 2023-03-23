package domain

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/jwt"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func Test_projectDomain_Create(t *testing.T) {
	Initialized()
	expiration := time.Minute
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	engine := jwt.NewEngine[model.AccessToken](config.TokenConfigs{
		Expiration: expiration,
		Secret:     "secret",
	})
	token, err := engine.Generate("", model.AccessToken{
		ID:      "id",
		Name:    "name",
		Address: "address",
	})
	assert.NoError(t, err)
	ginCtx.Header("Authorization", fmt.Sprintf("Bearer %s", token))
	ctx := &router.Context{
		Context: ginCtx,
	}
	req := &model.CreateProjectRequest{
		Name:     "test",
		Twitter:  "https://twitter.com/hashtag/Breaking2",
		Discord:  "https://discord.com/hashtag/Breaking2",
		Telegram: "https://telegram.com/",
	}
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
}
