package testutil

import (
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/xcontext"
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

	// Collaborators
	Collaborators = []*entity.Collaborator{
		{
			Base:      entity.Base{ID: "collaborator1"},
			ProjectID: Project1.ID,
			UserID:    User1.ID,
			Role:      entity.Owner,
		},
		{
			Base:      entity.Base{ID: "collaborator3"},
			ProjectID: Project1.ID,
			UserID:    User3.ID,
			CreatedBy: User1.ID,
			Role:      entity.Reviewer,
		},
	}

	Collaborator1 = Collaborators[0]
	Collaborator2 = Collaborators[1]

	// Quests
	Quests = []*entity.Quest{
		{
			Base: entity.Base{
				ID: "project1_quest1",
			},
			ProjectID:      Project1.ID,
			Type:           entity.Text,
			Status:         entity.QuestDraft,
			Title:          "Quest1",
			Description:    "Quest1 Description",
			CategoryIDs:    []string{"1", "2", "3"},
			Recurrence:     entity.Once,
			ValidationData: `{}`,
			Awards:         []entity.Award{{Type: "point", Value: "100"}},
			ConditionOp:    entity.Or,
			Conditions:     []entity.Condition{{Type: "quest", Op: "is_completed", Value: "project1_quest1"}},
		},
		{
			Base: entity.Base{
				ID: "project1_quest2",
			},
			ProjectID:      Project1.ID,
			Type:           entity.VisitLink,
			Status:         entity.QuestActive,
			Title:          "Quest2",
			Description:    "Quest2 Description",
			CategoryIDs:    []string{},
			Recurrence:     entity.Daily,
			ValidationData: `{"link": "https://example.com"}`,
			Awards:         []entity.Award{},
			ConditionOp:    entity.And,
			Conditions:     []entity.Condition{{Type: "quest", Op: "is_completed", Value: "project1_quest1"}},
		},
	}

	Quest1 = Quests[0]
	Quest2 = Quests[1]

	// Cateogories
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

	// ClaimedQuests
	ClaimedQuests = []*entity.ClaimedQuest{
		{
			Base:    entity.Base{ID: "claimedQuest1"},
			QuestID: Quest1.ID,
			UserID:  User1.ID,
			Status:  entity.Accepted,
			Input:   "any",
		},
		{
			Base:    entity.Base{ID: "claimedQuest2"},
			QuestID: Quest2.ID,
			UserID:  User2.ID,
			Status:  entity.Rejected,
			Input:   "bar",
		},
		{
			Base:    entity.Base{ID: "claimedQuest3"},
			QuestID: Quest2.ID,
			UserID:  User3.ID,
			Status:  entity.Pending,
			Input:   "foo",
		},
	}

	ClaimedQuest1 = ClaimedQuests[0]
	ClaimedQuest2 = ClaimedQuests[1]
	ClaimedQuest3 = ClaimedQuests[2]
)

func CreateFixtureDb(ctx xcontext.Context) {
	InsertUsers(ctx)
	InsertProjects(ctx)
	InsertCollaborators(ctx)
	InsertCategories(ctx)
	InsertQuests(ctx)
	InsertClaimedQuests(ctx)
}

func InsertUsers(ctx xcontext.Context) {
	var err error
	userRepo := repository.NewUserRepository()

	for _, user := range Users {
		err = userRepo.Create(ctx, user)
		if err != nil {
			panic(err)
		}
	}
}

func InsertProjects(ctx xcontext.Context) {
	projectRepo := repository.NewProjectRepository()

	for _, project := range Projects {
		err := projectRepo.Create(ctx, project)
		if err != nil {
			panic(err)
		}
	}
}

func InsertCollaborators(ctx xcontext.Context) {
	collaboratorRepo := repository.NewCollaboratorRepository()

	for _, collaborator := range Collaborators {
		err := collaboratorRepo.Create(ctx, collaborator)
		if err != nil {
			panic(err)
		}
	}
}

func InsertQuests(ctx xcontext.Context) {
	questRepo := repository.NewQuestRepository()

	for _, quest := range Quests {
		err := questRepo.Create(ctx, quest)
		if err != nil {
			panic(err)
		}
	}
}

func InsertCategories(ctx xcontext.Context) {
	categoryRepo := repository.NewCategoryRepository()

	for _, category := range Categories {
		err := categoryRepo.Create(ctx, category)
		if err != nil {
			panic(err)
		}
	}
}

func InsertClaimedQuests(ctx xcontext.Context) {
	claimedQuestRepo := repository.NewClaimedQuestRepository()

	for _, claimedQuest := range ClaimedQuests {
		err := claimedQuestRepo.Create(ctx, claimedQuest)
		if err != nil {
			panic(err)
		}
	}
}
