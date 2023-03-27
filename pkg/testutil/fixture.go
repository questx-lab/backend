package testutil

import (
	"context"

	"github.com/google/uuid"
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
		{
			Base: entity.Base{
				ID: "user3",
			},
		},
	}
	User1 = Users[0]
	User2 = Users[1]
	User3 = Users[2]

	// Projects
	Projects = []*entity.Project{
		{
			Base: entity.Base{
				ID: "user1_project1",
			},
			Name:      "User1 Project1",
			CreatedBy: User1.ID,
			Twitter:   "https://twitter.com/hashtag/Breaking2",
			Discord:   "https://discord.com/hashtag/Breaking2",
			Telegram:  "https://telegram.com",
		},
		{
			Base: entity.Base{
				ID: "user2_project2",
			},
			Name:      "User2 Project2",
			CreatedBy: User2.ID,
			Twitter:   "https://twitter.com/hashtag/Breaking2",
			Discord:   "https://discord.com/hashtag/Breaking2",
			Telegram:  "https://telegram.com",
		},
	}
	Project1 = Projects[0]
	Project2 = Projects[1]

	// Quests
	Quests = []*entity.Quest{
		{
			Base: entity.Base{
				ID: "project1_quest1",
			},
			ProjectID:      Project1.ID,
			Type:           entity.QuestText,
			Status:         entity.QuestStatusDraft,
			Title:          "Quest1",
			Description:    "Quest1 Description",
			CategoryIDs:    []string{"1", "2", "3"},
			Recurrence:     entity.QuestRecurrenceOnce,
			ValidationData: "",
			Awards:         []entity.Award{{Type: "point", Value: "100"}},
			ConditionOp:    entity.QuestConditionOpOr,
			Conditions:     []entity.Condition{{Type: "level", Op: "<=", Value: "15"}},
		},
		{
			Base: entity.Base{
				ID: "project1_quest2",
			},
			ProjectID:      Project1.ID,
			Type:           entity.QuestVisitLink,
			Status:         entity.QuestStatusPublished,
			Title:          "Quest2",
			Description:    "Quest2 Description",
			CategoryIDs:    []string{},
			Recurrence:     entity.QuestRecurrenceDaily,
			ValidationData: "",
			Awards:         []entity.Award{{Type: "discord role", Value: "mod"}},
			ConditionOp:    entity.QuestConditionOpAnd,
			Conditions:     []entity.Condition{},
		},
	}

	Quest1 = Quests[0]
	Quest2 = Quests[1]

	Categories = []*entity.Category{
		{
			Base:      entity.Base{ID: "category1"},
			Name:      "Category 1",
			ProjectID: Project1.ID,
			CreatedBy: User1.ID,
		},
		{
			Base:      entity.Base{ID: "category2"},
			Name:      "Category 2",
			ProjectID: Project1.ID,
			CreatedBy: User1.ID,
		},
		{
			Base:      entity.Base{ID: "category3"},
			Name:      "Category 3",
			ProjectID: Project1.ID,
			CreatedBy: User3.ID,
		},
	}

	Category1 = Categories[0]
	Category2 = Categories[1]
	Category3 = Categories[2]
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
	InsertCollaborators(db)
	InsertCategories(db)
	InsertQuests(db)

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

func InsertCollaborators(db *gorm.DB) {
	ctx := context.Background()
	collaboratorRepo := repository.NewCollaboratorRepository(db)

	c1 := &entity.Collaborator{
		Base:      entity.Base{ID: uuid.NewString()},
		ProjectID: Project1.ID,
		UserID:    User1.ID,
		CreatedBy: "valid-created-by",
		Role:      entity.Owner,
	}

	if err := collaboratorRepo.Create(ctx, c1); err != nil {
		panic(err)
	}

	c3 := &entity.Collaborator{
		Base:      entity.Base{ID: uuid.NewString()},
		ProjectID: Project1.ID,
		UserID:    User3.ID,
		CreatedBy: "valid-created-by",
		Role:      entity.Reviewer,
	}

	if err := collaboratorRepo.Create(ctx, c3); err != nil {
		panic(err)
	}
}

func InsertQuests(db *gorm.DB) {
	questRepo := repository.NewQuestRepository(db)

	for _, quest := range Quests {
		err := questRepo.Create(context.Background(), quest)
		if err != nil {
			panic(err)
		}
	}
}

func InsertCategories(db *gorm.DB) {
	categoryRepo := repository.NewCategoryRepository(db)

	for _, category := range Categories {
		err := categoryRepo.Create(context.Background(), category)
		if err != nil {
			panic(err)
		}
	}
}
