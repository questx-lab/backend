package domain

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/authenticator"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/testutil"

	"gorm.io/gorm"
)

func generateRandomString() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	state := base64.StdEncoding.EncodeToString(b)

	return state, nil
}

var (
	projectRepo   repository.ProjectRepository
	projectdomain ProjectDomain
	db            *gorm.DB
	validUserID   string
)

func Initialized() {
	db = testutil.GetDatabaseTest()
	projectRepo = repository.NewProjectRepository(db)
	projectdomain = NewProjectDomain(projectRepo)
	_ = db.AutoMigrate(&entity.Project{})
	validUserID = "valid-user-id"
}

func NewMockContext() router.Context {
	ctx := router.DefaultContext()
	r := httptest.NewRequest(http.MethodGet, "/createProject", nil)
	tokenEngine := authenticator.NewTokenEngine[model.AccessToken](config.TokenConfigs{
		Secret:     "secret",
		Expiration: time.Minute,
	})

	tkn, _ := tokenEngine.Generate(validUserID, model.AccessToken{
		ID: validUserID,
	})
	r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tkn))
	ctx.SetRequest(r)
	ctx.SetAccessTokenEngine(tokenEngine)
	return ctx
}
