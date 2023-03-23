package domain

import (
	"crypto/rand"
	"encoding/base64"
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
)

func Initialized() {
	db = testutil.GetDatabaseTest()
	projectRepo = repository.NewProjectRepository(db)
	projectdomain = NewProjectDomain(projectRepo)
	db.AutoMigrate(&entity.Project{})
}
