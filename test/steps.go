package test

import (
	"context"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
)

func (s *suite) createProject() error {
	ctx := context.Background()
	projectRepo := repository.NewProjectRepository(s.db)

	p := &entity.Project{
		Base:      entity.Base{ID: uuid.NewString()},
		Name:      uuid.NewString(),
		CreatedBy: uuid.NewString(),
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

func (s *suite) createCollaborator(role entity.CollaboratorRole) error {
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
