package testutil

import (
	"context"
	"database/sql"
	"math"

	"github.com/questx-lab/backend/internal/domain/badge"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
)

var (
	// Roles
	Roles = []*entity.Role{
		{
			Base: entity.Base{ID: "user"},
			Name: "user",
		},
		{
			Base:        entity.Base{ID: "editor"},
			Name:        "editor",
			Permissions: 7,
		},
		{
			Base:        entity.Base{ID: "reviewer"},
			Name:        "reviewer",
			Permissions: 8,
		},
		{
			Base:        entity.Base{ID: "owner"},
			Name:        "owner",
			Permissions: math.MaxInt64,
		},
	}

	// Users
	Users = []*entity.User{
		{
			Base:           entity.Base{ID: "user1"},
			Name:           "user1",
			Role:           entity.RoleSuperAdmin,
			ReferralCode:   "uHi0K",
			WalletAddress:  sql.NullString{Valid: true, String: "0x0000"},
			ProfilePicture: "https://x.com/avatar.jpg",
		},
		{
			Base:          entity.Base{ID: "user2"},
			Name:          "user2",
			WalletAddress: sql.NullString{Valid: true, String: "random-wallet-address"},
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
			Handle:      "user1_community1",
			DisplayName: "User1 Community1",
			CreatedBy:   User1.ID,
			Twitter:     "https://twitter.com/hashtag/Breaking2",
			Discord:     "1234",
		},
		{
			Base: entity.Base{
				ID: "user2_community2",
			},
			Handle:      "user2_community2",
			DisplayName: "User2 Community2",
			CreatedBy:   User2.ID,
			Twitter:     "https://twitter.com/hashtag/Breaking2",
			Discord:     "5678",
		},
	}
	Community1 = Communities[0]
	Community2 = Communities[1]

	// Followers
	Followers = []*entity.Follower{
		{
			UserID:      User1.ID,
			CommunityID: Community1.ID,
			InviteCode:  "Foo",
			Points:      1000,
			Quests:      10,
			RoleID:      "owner",
		},
		{
			UserID:      User2.ID,
			CommunityID: Community2.ID,
			InviteCode:  "Foo Foo 2",
			Points:      1000,
			Quests:      10,
			RoleID:      "owner",
		},
		{
			UserID:      User1.ID,
			CommunityID: Community2.ID,
			InviteCode:  "Foo Foo",
			Points:      1000,
			Quests:      10,
			RoleID:      "user",
		},
		{
			UserID:      User2.ID,
			CommunityID: Community1.ID,
			InviteCode:  "Bar",
			Points:      1000,
			Quests:      10,
			RoleID:      "user",
		},
		{
			UserID:      User3.ID,
			CommunityID: Community1.ID,
			InviteCode:  "Far",
			Points:      1000,
			Quests:      10,
			RoleID:      "editor",
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
			Base:           entity.Base{ID: "claimedQuest1"},
			QuestID:        Quest1.ID,
			UserID:         User1.ID,
			Status:         entity.Accepted,
			SubmissionData: "any",
		},
		{
			Base:           entity.Base{ID: "claimedQuest2"},
			QuestID:        Quest2.ID,
			UserID:         User2.ID,
			Status:         entity.Rejected,
			SubmissionData: "bar",
		},
		{
			Base:           entity.Base{ID: "claimedQuest3"},
			QuestID:        Quest2.ID,
			UserID:         User3.ID,
			Status:         entity.Pending,
			SubmissionData: "foo",
		},
		{
			Base:           entity.Base{ID: "claimedQuest4"},
			QuestID:        Quest2.ID,
			UserID:         User1.ID,
			Status:         entity.Accepted,
			SubmissionData: "any",
		},
	}

	ClaimedQuest1 = ClaimedQuests[0]
	ClaimedQuest2 = ClaimedQuests[1]
	ClaimedQuest3 = ClaimedQuests[2]

	Badges = []entity.Badge{
		{
			Base:  entity.Base{ID: "badge_1"},
			Name:  badge.SharpScoutBadgeName,
			Level: 1,
			Value: 1,
		},
		{
			Base:  entity.Base{ID: "badge_2"},
			Name:  badge.SharpScoutBadgeName,
			Level: 2,
			Value: 2,
		},
		{
			Base:  entity.Base{ID: "badge_3"},
			Name:  badge.SharpScoutBadgeName,
			Level: 3,
			Value: 3,
		},
		{
			Base:  entity.Base{ID: "badge_4"},
			Name:  badge.QuestWarriorBadgeName,
			Level: 1,
			Value: 1,
		},
		{
			Base:  entity.Base{ID: "badge_5"},
			Name:  badge.QuestWarriorBadgeName,
			Level: 2,
			Value: 2,
		},
		{
			Base:  entity.Base{ID: "badge_6"},
			Name:  badge.QuestWarriorBadgeName,
			Level: 3,
			Value: 3,
		},
		{
			Base:  entity.Base{ID: "badge_7"},
			Name:  badge.RainBowBadgeName,
			Level: 1,
			Value: 1,
		},
		{
			Base:  entity.Base{ID: "badge_8"},
			Name:  badge.RainBowBadgeName,
			Level: 2,
			Value: 2,
		},
		{
			Base:  entity.Base{ID: "badge_9"},
			Name:  badge.RainBowBadgeName,
			Level: 3,
			Value: 3,
		},
	}

	BadgeSharpScout1   = Badges[0]
	BadgeSharpScout2   = Badges[1]
	BadgeSharpScout3   = Badges[2]
	BadgeQuestWarrior1 = Badges[3]
	BadgeQuestWarrior2 = Badges[4]
	BadgeQuestWarrior3 = Badges[5]
	BadgeRainbow1      = Badges[6]
	BadgeRainbow2      = Badges[7]
	BadgeRainbow3      = Badges[8]
)

func CreateFixtureDb(ctx context.Context) {
	InsertUsers(ctx)
	InsertRoles(ctx)
	InsertCommunities(ctx)
	InsertFollowers(ctx)
	InsertCategories(ctx)
	InsertQuests(ctx)
	InsertClaimedQuests(ctx)
	InsertBadges(ctx)
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

func InsertRoles(ctx context.Context) {
	var err error
	roleRepo := repository.NewRoleRepository()

	for _, role := range Roles {
		err = roleRepo.Create(ctx, role)
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

func InsertBadges(ctx context.Context) {
	badgeRepo := repository.NewBadgeRepository()

	for _, badge := range Badges {
		err := badgeRepo.Create(ctx, &badge)
		if err != nil {
			panic(err)
		}
	}
}
