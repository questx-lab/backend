package test

import (
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/testutil"

	"gorm.io/gorm"
)

type suite struct {
	db *gorm.DB

	Project      *entity.Project
	User         *entity.User
	Collaborator *entity.Collaborator
}

func NewSuite() *suite {
	return &suite{
		db: testutil.GetDatabaseTest(),
	}
}
