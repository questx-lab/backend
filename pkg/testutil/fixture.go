package testutil

import (
	"context"
	"log"
	"os"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	FixtureDb = "test.db"
)

func CreateFixture() {
	// Write to bak file first in case this fails.
	db, err := gorm.Open(sqlite.Open(FixtureDb+".bak"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	if err := entity.MigrateTable(db); err != nil {
		panic(err)
	}

	InsertUser(db)
	InsertProject(db)

	// Rename bak file
	e := os.Rename(FixtureDb+".bak", FixtureDb)
	if e != nil {
		log.Fatal(e)
	}
}

func InsertUser(db *gorm.DB) {
	var err error
	userRepo := repository.NewUserRepository(db)

	// user1
	err = userRepo.Create(context.Background(), &entity.User{
		Base: entity.Base{
			ID: "user1",
		},
	})
	if err != nil {
		panic(err)
	}

	// user2
	err = userRepo.Create(context.Background(), &entity.User{
		Base: entity.Base{
			ID: "user2",
		},
	})
	if err != nil {
		panic(err)
	}
}

func InsertProject(db *gorm.DB) {
	projectRepo := repository.NewProjectRepository(db)
	err := projectRepo.Create(context.Background(), &entity.Project{
		Base: entity.Base{
			ID: "user1_project1",
		},
		Name:      "User1 Project1",
		CreatedBy: "user1",
	})

	if err != nil {
		panic(err)
	}
}
