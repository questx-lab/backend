package domain

import (
	"context"
	"testing"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/testutil"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type suite struct {
	db *gorm.DB

	Project      *entity.Project
	User         *entity.User
	Collaborator *entity.Collaborator
}

func NewSuite(t *testing.T) *suite {
	return &suite{
		db: testutil.DefaultTestDb(t),
	}
}

func (s *suite) createProject() error {
	ctx := context.Background()
	projectRepo := repository.NewProjectRepository(s.db)

	p := &entity.Project{
		Base:      entity.Base{ID: uuid.NewString()},
		Name:      "valid-project",
		CreatedBy: s.User.ID,
		Twitter:   "https://twitter.com/hashtag/Breaking2",
		Discord:   "https://discord.com/hashtag/Breaking2",
		Telegram:  "https://telegram.com",
	}

	if err := projectRepo.Create(ctx, p); err != nil {
		return err
	}
	s.Project = p
	return nil
}

func (s *suite) createUser() error {
	ctx := context.Background()
	userRepo := repository.NewUserRepository(s.db)

	u := &entity.User{
		Base:    entity.Base{ID: uuid.NewString()},
		Name:    "valid-user",
		Address: "valid-address",
	}

	if err := userRepo.Create(ctx, u); err != nil {
		return err
	}
	s.User = u

	return nil
}

func (s *suite) createCollaborator(role entity.Role) error {
	ctx := context.Background()
	collaboratorRepo := repository.NewCollaboratorRepository(s.db)

	c := &entity.Collaborator{
		Base:      entity.Base{ID: uuid.NewString()},
		ProjectID: s.Project.ID,
		UserID:    s.User.ID,
		CreatedBy: "valid-created-by",
		Role:      role,
	}

	if err := collaboratorRepo.Create(ctx, c); err != nil {
		return err
	}

	s.Collaborator = c

	return nil
}

func (s *suite) updateCollaboratorRole(role entity.Role) error {
	ctx := context.Background()
	collaboratorRepo := repository.NewCollaboratorRepository(s.db)

	if err := collaboratorRepo.UpdateRole(ctx, s.User.ID, s.Project.ID, role); err != nil {
		return err
	}

	return nil
}
