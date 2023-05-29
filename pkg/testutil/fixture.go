package testutil

import (
	"context"
	"database/sql"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
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
			Base:    entity.Base{ID: "user2"},
			Name:    "user2",
			Address: sql.NullString{Valid: true, String: "random-wallet-address"},
		},
		{
			Base: entity.Base{ID: "user3"},
			Name: "user3",
		},
	}
	User1 = Users[0]
	User2 = Users[1]
	User3 = Users[2]

	// Communities
	Communities = []*entity.Community{
		{
			Base: entity.Base{
				ID: "user1_community1",
			},
			Handle:      "User1_Community1",
			DisplayName: "User1 Community1",
			CreatedBy:   User1.ID,
			Twitter:     "https://twitter.com/hashtag/Breaking2",
			Discord:     "1234",
		},
		{
			Base: entity.Base{
				ID: "user2_community2",
			},
			Handle:      "User2_Community2",
			DisplayName: "User2 Community2",
			CreatedBy:   User2.ID,
			Twitter:     "https://twitter.com/hashtag/Breaking2",
			Discord:     "5678",
		},
	}
	Community1 = Communities[0]
	Community2 = Communities[1]

	// Collaborators
	Collaborators = []*entity.Collaborator{
		{
			CommunityID: Community1.ID,
			UserID:      Community1.CreatedBy,
			Role:        entity.Owner,
			CreatedBy:   Community1.CreatedBy,
		},
		{
			CommunityID: Community2.ID,
			UserID:      Community2.CreatedBy,
			Role:        entity.Owner,
			CreatedBy:   Community2.CreatedBy,
		},
		{
			CommunityID: Community1.ID,
			UserID:      User3.ID,
			CreatedBy:   User1.ID,
			Role:        entity.Reviewer,
		},
	}

	Collaborator1 = Collaborators[0]
	Collaborator2 = Collaborators[1]
	Collaborator3 = Collaborators[2]

	// Followers
	Followers = []*entity.Follower{
		{
			UserID:      User1.ID,
			CommunityID: Community1.ID,
			InviteCode:  "Foo",
		},
		{
			UserID:      User2.ID,
			CommunityID: Community1.ID,
			InviteCode:  "Bar",
		},
		{
			UserID:      User3.ID,
			CommunityID: Community1.ID,
			InviteCode:  "Far",
		},
	}

	Follower1 = Followers[0]
	Follower2 = Followers[1]
	Follower3 = Followers[2]

	// Quests
	Quests = []*entity.Quest{
		{
			Base: entity.Base{
				ID: "community1_quest1",
			},
			CommunityID:    sql.NullString{Valid: true, String: Community1.ID},
			Type:           entity.QuestText,
			Status:         entity.QuestDraft,
			Title:          "Quest1",
			Description:    []byte("Quest1 Description"),
			CategoryID:     sql.NullString{Valid: true, String: "category1"},
			Recurrence:     entity.Once,
			ValidationData: entity.Map{},
			Points:         100,
			ConditionOp:    entity.Or,
		},
		{
			Base: entity.Base{
				ID: "community1_quest2",
			},
			CommunityID:    sql.NullString{Valid: true, String: Community1.ID},
			Type:           entity.QuestVisitLink,
			Status:         entity.QuestActive,
			Title:          "Quest2",
			Description:    []byte("Quest2 Description"),
			Recurrence:     entity.Daily,
			ValidationData: entity.Map{"link": "https://example.com"},
			Points:         100,
			ConditionOp:    entity.And,
			Conditions: []entity.Condition{
				{
					Type: "quest",
					Data: entity.Map{"op": "is_completed", "quest_title": "Quest 1", "quest_id": "community1_quest1"},
				},
			},
		},
		{
			Base: entity.Base{
				ID: "community1_quest3",
			},
			CommunityID:    sql.NullString{Valid: true, String: Community1.ID},
			Type:           entity.QuestVisitLink,
			Status:         entity.QuestActive,
			Title:          "Quest3",
			Description:    []byte("Quest2 Description"),
			Recurrence:     entity.Daily,
			ValidationData: entity.Map{"link": "https://example.com"},
			Points:         100,
			ConditionOp:    entity.And,
			Conditions:     []entity.Condition{},
		},
		{
			Base: entity.Base{
				ID: "template_quest4",
			},
			CommunityID:    sql.NullString{Valid: false},
			IsTemplate:     true,
			Type:           entity.QuestText,
			Status:         entity.QuestDraft,
			Title:          "Quest of {{ .community.DisplayName }}",
			Description:    []byte("Description is written by {{ .owner.Name }} for {{ .community.DisplayName }}"),
			Recurrence:     entity.Once,
			ValidationData: entity.Map{},
			Points:         100,
			ConditionOp:    entity.Or,
		},
	}

	Quest1        = Quests[0]
	Quest2        = Quests[1]
	Quest3        = Quests[2]
	QuestTemplate = Quests[3]

	// Cateogories
	Categories = []*entity.Category{
		{
			Base:        entity.Base{ID: "category1"},
			Name:        "Category 1",
			CommunityID: sql.NullString{Valid: true, String: Community1.ID},
			CreatedBy:   User1.ID,
		},
		{
			Base:        entity.Base{ID: "category2"},
			Name:        "Category 2",
			CommunityID: sql.NullString{Valid: true, String: Community1.ID},
			CreatedBy:   User1.ID,
		},
		{
			Base:        entity.Base{ID: "category3"},
			Name:        "Category 3",
			CommunityID: sql.NullString{Valid: true, String: Community1.ID},
			CreatedBy:   User3.ID,
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

func CreateFixtureDb(ctx context.Context) {
	InsertUsers(ctx)
	InsertCommunities(ctx)
	InsertFollowers(ctx)
	InsertCollaborators(ctx)
	InsertCategories(ctx)
	InsertQuests(ctx)
	InsertClaimedQuests(ctx)
}

func InsertUsers(ctx context.Context) {
	var err error
	userRepo := repository.NewUserRepository()

	for _, user := range Users {
		err = userRepo.Create(ctx, user)
		if err != nil {
			panic(err)
		}
	}
}

func InsertCommunities(ctx context.Context) {
	communityRepo := repository.NewCommunityRepository(&MockSearchCaller{})

	for _, community := range Communities {
		err := communityRepo.Create(ctx, community)
		if err != nil {
			panic(err)
		}
	}
}

func InsertFollowers(ctx context.Context) {
	followerRepo := repository.NewFollowerRepository()

	for _, follower := range Followers {
		err := followerRepo.Create(ctx, follower)
		if err != nil {
			panic(err)
		}
	}
}

func InsertCollaborators(ctx context.Context) {
	collaboratorRepo := repository.NewCollaboratorRepository()

	for _, collaborator := range Collaborators {
		err := collaboratorRepo.Upsert(ctx, collaborator)
		if err != nil {
			panic(err)
		}
	}
}

func InsertQuests(ctx context.Context) {
	questRepo := repository.NewQuestRepository(&MockSearchCaller{})

	for _, quest := range Quests {
		err := questRepo.Create(ctx, quest)
		if err != nil {
			panic(err)
		}
	}
}

func InsertCategories(ctx context.Context) {
	categoryRepo := repository.NewCategoryRepository()

	for _, category := range Categories {
		err := categoryRepo.Create(ctx, category)
		if err != nil {
			panic(err)
		}
	}
}

func InsertClaimedQuests(ctx context.Context) {
	claimedQuestRepo := repository.NewClaimedQuestRepository()

	for _, claimedQuest := range ClaimedQuests {
		err := claimedQuestRepo.Create(ctx, claimedQuest)
		if err != nil {
			panic(err)
		}
	}
}
