package testutil

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	DbDump = "test/testdb.dump"
)

var (
	// Users
	Users = []*entity.User{
		{
			Base: entity.Base{
				ID: "user1",
			},
		},
		{
			Base: entity.Base{
				ID: "user2",
			},
		},
	}
	User1 = Users[0]
	User2 = Users[1]

	// Projects
	Projects = []*entity.Project{
		{
			Base: entity.Base{
				ID: "user1_project1",
			},
			Name:      "User1 Project1",
			CreatedBy: "user1",
			Twitter:   "https://twitter.com/hashtag/Breaking2",
			Discord:   "https://discord.com/hashtag/Breaking2",
			Telegram:  "https://telegram.com",
		},
	}
	Project1 = Projects[0]
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

	for _, user := range Users {
		err = userRepo.Create(context.Background(), user)
		if err != nil {
			panic(err)
		}
	}
}

func InsertProjects(db *gorm.DB) {
	projectRepo := repository.NewProjectRepository(db)

	for _, project := range Projects {
		err := projectRepo.Create(context.Background(), project)
		if err != nil {
			panic(err)
		}
	}
}
