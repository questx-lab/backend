package main

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"gorm.io/gorm"
)

func createFixture(db *gorm.DB) {
	if err := entity.MigrateTable(db); err != nil {
		panic(err)
	}

	projectRepo := repository.NewProjectRepository(db)
	userRepo := repository.NewUserRepository(db)

	userRepo.Create(context.Background(), &entity.User{
		Base: entity.Base{
			ID: "user1",
		},
	})

	projectRepo.Create(context.Background(), &entity.Project{
		Base: entity.Base{
			ID: "user1_project1",
		},
		Name:      "User1 Project1",
		CreatedBy: "user1",
	})
}
