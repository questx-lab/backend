package testutil

import (
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/dateutil"
	"github.com/questx-lab/backend/pkg/xcontext"
)

var (
	// Users
	Users = []*entity.User{
		{
			Base: entity.Base{ID: "user1"},
			Name: "user1",
			Role: entity.RoleSuperAdmin,
		},
		{
			Base: entity.Base{ID: "user2"},
			Name: "user2",
		},
		{
			Base: entity.Base{ID: "user3"},
			Name: "user3",
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
			Discord:   "1234",
		},
		{
			Base: entity.Base{
				ID: "user2_project2",
			},
			Name:      "User2 Project2",
			CreatedBy: User2.ID,
			Twitter:   "https://twitter.com/hashtag/Breaking2",
			Discord:   "5678",
		},
	}
	Project1 = Projects[0]
	Project2 = Projects[1]

	// Collaborators
	Collaborators = []*entity.Collaborator{
		{
			ProjectID: Project1.ID,
			UserID:    Project1.CreatedBy,
			Role:      entity.Owner,
			CreatedBy: Project1.CreatedBy,
		},
		{
			ProjectID: Project2.ID,
			UserID:    Project2.CreatedBy,
			Role:      entity.Owner,
			CreatedBy: Project2.CreatedBy,
		},
		{
			ProjectID: Project1.ID,
			UserID:    User3.ID,
			CreatedBy: User1.ID,
			Role:      entity.Reviewer,
		},
	}

	Collaborator1 = Collaborators[0]
	Collaborator2 = Collaborators[1]
	Collaborator3 = Collaborators[2]

	// Participants
	Participants = []*entity.Participant{
		{
			UserID:     User1.ID,
			ProjectID:  Project1.ID,
			InviteCode: "Foo",
		},
		{
			UserID:     User2.ID,
			ProjectID:  Project1.ID,
			InviteCode: "Bar",
		},
		{
			UserID:     User3.ID,
			ProjectID:  Project1.ID,
			InviteCode: "Far",
		},
	}

	Participant1 = Participants[0]
	Participant2 = Participants[1]
	Participant3 = Participants[2]

	// Quests
	Quests = []*entity.Quest{
		{
			Base: entity.Base{
				ID: "project1_quest1",
			},
			ProjectID:      Project1.ID,
			Type:           entity.QuestText,
			Status:         entity.QuestDraft,
			Title:          "Quest1",
			Description:    []byte("Quest1 Description"),
			CategoryIDs:    []string{"1", "2", "3"},
			Recurrence:     entity.Once,
			ValidationData: entity.Map{},
			Rewards:        []entity.Reward{{Type: "points", Data: entity.Map{"points": 100}}},
			ConditionOp:    entity.Or,
		},
		{
			Base: entity.Base{
				ID: "project1_quest2",
			},
			ProjectID:      Project1.ID,
			Type:           entity.QuestVisitLink,
			Status:         entity.QuestActive,
			Title:          "Quest2",
			Description:    []byte("Quest2 Description"),
			CategoryIDs:    []string{},
			Recurrence:     entity.Daily,
			ValidationData: entity.Map{"link": "https://example.com"},
			Rewards:        []entity.Reward{{Type: "points", Data: entity.Map{"points": 100}}},
			ConditionOp:    entity.And,
			Conditions: []entity.Condition{
				{
					Type: "quest",
					Data: entity.Map{"op": "is_completed", "quest_title": "Quest 1", "quest_id": "project1_quest1"},
				},
			},
		},
		{
			Base: entity.Base{
				ID: "project1_quest3",
			},
			ProjectID:      Project1.ID,
			Type:           entity.QuestVisitLink,
			Status:         entity.QuestActive,
			Title:          "Quest3",
			Description:    []byte("Quest2 Description"),
			CategoryIDs:    []string{},
			Recurrence:     entity.Daily,
			ValidationData: entity.Map{"link": "https://example.com"},
			Rewards:        []entity.Reward{{Type: "points", Data: entity.Map{"points": 100}}},
			ConditionOp:    entity.And,
			Conditions:     []entity.Condition{},
		},
	}

	Quest1 = Quests[0]
	Quest2 = Quests[1]
	Quest3 = Quests[2]

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

	aVal, _    = dateutil.GetCurrentValueByRange(entity.UserAggregateRangeWeek)
	prevVal, _ = dateutil.GetValueByRange(time.Now().AddDate(0, 0, -7), entity.UserAggregateRangeWeek)

	UserAggregates = []*entity.UserAggregate{
		{
			ProjectID:  Project2.ID,
			UserID:     User1.ID,
			RangeValue: aVal,
			Range:      entity.UserAggregateRangeWeek,
			TotalTask:  1,
			TotalPoint: 3,
		},
		{
			ProjectID:  Project2.ID,
			UserID:     User2.ID,
			RangeValue: aVal,
			Range:      entity.UserAggregateRangeWeek,
			TotalTask:  2,
			TotalPoint: 2,
		},
		{
			ProjectID:  Project2.ID,
			UserID:     User3.ID,
			RangeValue: aVal,
			Range:      entity.UserAggregateRangeWeek,
			TotalTask:  3,
			TotalPoint: 1,
		},

		// prev week
		{
			ProjectID:  Project2.ID,
			UserID:     User1.ID,
			RangeValue: prevVal,
			Range:      entity.UserAggregateRangeWeek,
			TotalTask:  1,
			TotalPoint: 3,
		},
		{
			ProjectID:  Project2.ID,
			UserID:     User2.ID,
			RangeValue: prevVal,
			Range:      entity.UserAggregateRangeWeek,
			TotalTask:  2,
			TotalPoint: 2,
		},
		{
			ProjectID:  Project2.ID,
			UserID:     User3.ID,
			RangeValue: prevVal,
			Range:      entity.UserAggregateRangeWeek,
			TotalTask:  0,
			TotalPoint: 0,
		},
	}

	UserAggregate1 = UserAggregates[0]
	UserAggregate2 = UserAggregates[1]
	UserAggregate3 = UserAggregates[2]
)

func CreateFixtureDb(ctx xcontext.Context) {
	InsertUsers(ctx)
	InsertProjects(ctx)
	InsertParticipants(ctx)
	InsertCollaborators(ctx)
	InsertCategories(ctx)
	InsertQuests(ctx)
	InsertClaimedQuests(ctx)
	InsertUserAggregates(ctx)
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

func InsertParticipants(ctx xcontext.Context) {
	participantRepo := repository.NewParticipantRepository()

	for _, participant := range Participants {
		err := participantRepo.Create(ctx, participant)
		if err != nil {
			panic(err)
		}
	}
}

func InsertCollaborators(ctx xcontext.Context) {
	collaboratorRepo := repository.NewCollaboratorRepository()

	for _, collaborator := range Collaborators {
		err := collaboratorRepo.Upsert(ctx, collaborator)
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

func InsertUserAggregates(ctx xcontext.Context) {
	achievementRepo := repository.NewUserAggregateRepository()
	for _, ua := range UserAggregates {
		if err := achievementRepo.Upsert(ctx, ua); err != nil {
			panic(err)
		}
	}
}
