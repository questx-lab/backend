package testutil

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	DbDump = "testdb.dump"
)

func CreateFixtureDb() *gorm.DB {
	// 1. Create in memory db
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	if err := entity.MigrateTable(db); err != nil {
		panic(err)
	}

	// 2. Insert data
	InsertUsers(db)
	InsertProjects(db)

	return db
}

func InsertUsers(db *gorm.DB) {
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

func InsertProjects(db *gorm.DB) {
	projectRepo := repository.NewProjectRepository(db)
	err := projectRepo.Create(context.Background(), &entity.Project{
		Base: entity.Base{
			ID: "user1_project1",
		},
		Name:      "User1 Project1",
		CreatedBy: "user1",
		Twitter:   "https://twitter.com/hashtag/Breaking2",
		Discord:   "https://discord.com/hashtag/Breaking2",
		Telegram:  "https://telegram.com",
	})
	if err != nil {
		panic(err)
	}
}
