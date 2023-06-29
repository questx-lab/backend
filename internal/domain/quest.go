package domain

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/fatih/structs"
	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/questclaim"
	"github.com/questx-lab/backend/internal/domain/statistic"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/api/discord"
	"github.com/questx-lab/backend/pkg/api/telegram"
	"github.com/questx-lab/backend/pkg/api/twitter"
	"github.com/questx-lab/backend/pkg/enum"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/pubsub"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type QuestDomain interface {
	Create(context.Context, *model.CreateQuestRequest) (*model.CreateQuestResponse, error)
	Update(context.Context, *model.UpdateQuestRequest) (*model.UpdateQuestResponse, error)
	Get(context.Context, *model.GetQuestRequest) (*model.GetQuestResponse, error)
	GetList(context.Context, *model.GetListQuestRequest) (*model.GetListQuestResponse, error)
	Delete(context.Context, *model.DeleteQuestRequest) (*model.DeleteQuestResponse, error)
	GetTemplates(context.Context, *model.GetQuestTemplatesRequest) (*model.GetQuestTemplatestResponse, error)
	ParseTemplate(context.Context, *model.ParseQuestTemplatesRequest) (*model.ParseQuestTemplatestResponse, error)
}

type questDomain struct {
	questRepo        repository.QuestRepository
	communityRepo    repository.CommunityRepository
	categoryRepo     repository.CategoryRepository
	claimedQuestRepo repository.ClaimedQuestRepository
	userRepo         repository.UserRepository
	followerRepo     repository.FollowerRepository
	roleVerifier     *common.CommunityRoleVerifier
	questFactory     questclaim.Factory
	leaderboard      statistic.Leaderboard
}

func NewQuestDomain(
	questRepo repository.QuestRepository,
	communityRepo repository.CommunityRepository,
	categoryRepo repository.CategoryRepository,
	collaboratorRepo repository.CollaboratorRepository,
	userRepo repository.UserRepository,
	claimedQuestRepo repository.ClaimedQuestRepository,
	oauth2Repo repository.OAuth2Repository,
	payRewardRepo repository.PayRewardRepository,
	followerRepo repository.FollowerRepository,
	twitterEndpoint twitter.IEndpoint,
	discordEndpoint discord.IEndpoint,
	telegramEndpoint telegram.IEndpoint,
	leaderboard statistic.Leaderboard,
	publisher pubsub.Publisher,
) *questDomain {
	roleVerifier := common.NewCommunityRoleVerifier(collaboratorRepo, userRepo)

	return &questDomain{
		questRepo:        questRepo,
		communityRepo:    communityRepo,
		categoryRepo:     categoryRepo,
		claimedQuestRepo: claimedQuestRepo,
		userRepo:         userRepo,
		followerRepo:     followerRepo,
		roleVerifier:     common.NewCommunityRoleVerifier(collaboratorRepo, userRepo),
		leaderboard:      leaderboard,
		questFactory: questclaim.NewFactory(
			claimedQuestRepo,
			questRepo,
			communityRepo,
			nil, // No need to know follower information when creating quest.
			oauth2Repo,
			userRepo,
			payRewardRepo,
			roleVerifier,
			twitterEndpoint,
			discordEndpoint,
			telegramEndpoint,
			publisher,
		),
	}
}

func (d *questDomain) Create(
	ctx context.Context, req *model.CreateQuestRequest,
) (*model.CreateQuestResponse, error) {
	communityID := ""
	if req.CommunityHandle != "" {
		community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorx.New(errorx.NotFound, "Not found community")
			}

			xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
			return nil, errorx.Unknown
		}
		communityID = community.ID
	}

	if err := d.roleVerifier.Verify(ctx, communityID, entity.AdminGroup...); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	quest := &entity.Quest{
		Base:        entity.Base{ID: uuid.NewString()},
		CommunityID: sql.NullString{Valid: true, String: communityID},
		IsTemplate:  false,
		Title:       req.Title,
		Description: []byte(req.Description),
		IsHighlight: req.IsHighlight,
		Points:      req.Points,
	}

	if communityID == "" {
		quest.CommunityID = sql.NullString{Valid: false}
		quest.IsTemplate = true
	}

	var err error
	quest.Type, err = enum.ToEnum[entity.QuestType](req.Type)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid quest type: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid quest type %s", req.Type)
	}

	quest.Status, err = enum.ToEnum[entity.QuestStatusType](req.Status)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid quest status: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid quest status %s", req.Status)
	}

	quest.Recurrence, err = enum.ToEnum[entity.RecurrenceType](req.Recurrence)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid recurrence: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid recurrence %s", req.Recurrence)
	}

	quest.ConditionOp, err = enum.ToEnum[entity.ConditionOpType](req.ConditionOp)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid condition op: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid condition op %s", req.ConditionOp)
	}

	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create quest factory with user: %v", err)
		return nil, errorx.Unknown
	}

	for _, r := range req.Rewards {
		rType, err := enum.ToEnum[entity.RewardType](r.Type)
		if err != nil {
			xcontext.Logger(ctx).Debugf("Invalid reward type: %v", err)
			continue
		}

		reward, err := d.questFactory.NewReward(ctx, *quest, rType, r.Data)
		if err != nil {
			return nil, err
		}

		quest.Rewards = append(quest.Rewards, entity.Reward{Type: rType, Data: structs.Map(reward)})
	}

	for _, c := range req.Conditions {
		ctype, err := enum.ToEnum[entity.ConditionType](c.Type)
		if err != nil {
			return nil, errorx.New(errorx.BadRequest, "Invalid condition type %s", c.Type)
		}

		condition, err := d.questFactory.NewCondition(ctx, ctype, c.Data)
		if err != nil {
			return nil, err
		}

		quest.Conditions = append(quest.Conditions, entity.Condition{Type: ctype, Data: structs.Map(condition)})
	}

	var processor questclaim.Processor
	if communityID != "" {
		processor, err = d.questFactory.NewProcessor(ctx, *quest, req.ValidationData)
	} else {
		processor, err = d.questFactory.LoadProcessor(ctx, true, *quest, req.ValidationData)
	}
	if err != nil {
		return nil, err
	}
	quest.ValidationData = structs.Map(processor)

	if req.CategoryID != "" {
		quest.CategoryID = sql.NullString{Valid: true, String: req.CategoryID}
		category, err := d.categoryRepo.GetByID(ctx, req.CategoryID)
		if err != nil {
			xcontext.Logger(ctx).Debugf("Invalid category: %v", err)
			return nil, errorx.New(errorx.NotFound, "Invalid category")
		}

		if category.CommunityID.String != communityID {
			return nil, errorx.New(errorx.BadRequest, "Category doesn't belong to community")
		}
	}

	err = d.questRepo.Create(ctx, quest)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create quest: %v", err)
		return nil, errorx.Unknown
	}

	return &model.CreateQuestResponse{
		ID: quest.ID,
	}, nil
}

func (d *questDomain) Get(ctx context.Context, req *model.GetQuestRequest) (*model.GetQuestResponse, error) {
	if req.IncludeUnclaimableReason && xcontext.RequestUserID(ctx) == "" {
		return nil, errorx.New(errorx.Unauthenticated,
			"Need authenticated if include_unclaimable_reason is turned on")
	}

	if req.ID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty id")
	}

	quest, err := d.questRepo.GetByID(ctx, req.ID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get quest: %v", err)
		return nil, errorx.Unknown
	}

	includeSecret := false
	var community *entity.Community
	if quest.CommunityID.Valid {
		var err error
		community, err = d.communityRepo.GetByID(ctx, quest.CommunityID.String)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorx.New(errorx.NotFound, "Not found community")
			}

			xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
			return nil, errorx.Unknown
		}

		if req.EditMode {
			if err := d.roleVerifier.Verify(ctx, community.ID, entity.AdminGroup...); err != nil {
				xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
				return nil, errorx.New(errorx.PermissionDenied, "Only owner or editor can edit quest")
			}

			includeSecret = true
		}
	} else {
		// In case this is a quest template (no community id), we will always
		// return a full information response, no need to hide any information.
		includeSecret = true
	}

	if err := processValidationData(ctx, d.questFactory, includeSecret, quest); err != nil {
		return nil, err
	}

	var category *entity.Category
	if quest.CategoryID.Valid {
		category, err = d.categoryRepo.GetByID(ctx, quest.CategoryID.String)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get category: %v", err)
			return nil, errorx.Unknown
		}
	}

	resp := model.GetQuestResponse(
		convertQuest(quest, convertCommunity(community, 0), convertCategory(category)))

	if req.IncludeUnclaimableReason {
		reason, err := d.questFactory.IsClaimable(ctx, *quest)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot determine not claimable reason: %v", err)
			return nil, errorx.Unknown
		}

		if reason != nil {
			resp.UnclaimableReason = reason.Message
		}
	}

	return &resp, nil
}

func (d *questDomain) GetList(
	ctx context.Context, req *model.GetListQuestRequest,
) (*model.GetListQuestResponse, error) {
	if req.IncludeUnclaimableReason && xcontext.RequestUserID(ctx) == "" {
		return nil, errorx.New(errorx.Unauthenticated,
			"Need authenticated if include_unclaimable_reason is turned on")
	}

	// No need to bound the limit parameter because the number of quests is
	// usually small. Moreover, the frontend can get all quests to allow user
	// searching quests.
	// If the limit is not set, this method will return all quests by default.
	if req.Limit == 0 {
		req.Limit = -1
	}

	var communityID string
	if req.CommunityHandle != "" {
		community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorx.New(errorx.NotFound, "Not found community")
			}

			xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
			return nil, errorx.Unknown
		}

		communityID = community.ID
	}

	categoryIDs := []string{}
	if req.CategoryIDs != "" {
		categoryIDs = strings.Split(req.CategoryIDs, ",")
	}

	statuses := []entity.QuestStatusType{entity.QuestActive, entity.QuestArchived}
	if communityID != "" {
		if d.roleVerifier.Verify(ctx, communityID, entity.AdminGroup...) == nil {
			statuses = append(statuses, entity.QuestDraft)
		}
	}

	quests, err := d.questRepo.GetList(ctx, repository.SearchQuestFilter{
		Q:           req.Q,
		CommunityID: communityID,
		CategoryIDs: categoryIDs,
		Offset:      req.Offset,
		Limit:       req.Limit,
		Statuses:    statuses,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get list of quests: %v", err)
		return nil, errorx.Unknown
	}

	categories, err := d.categoryRepo.GetList(ctx, communityID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get category: %v", err)
		return nil, errorx.Unknown
	}

	categoryMap := map[string]*entity.Category{}
	for i := range categories {
		categoryMap[categories[i].ID] = &categories[i]
	}

	communityMap := map[string]*entity.Community{}
	for i := range quests {
		if quests[i].CommunityID.Valid {
			communityMap[quests[i].CommunityID.String] = nil
		}
	}

	communities, err := d.communityRepo.GetByIDs(ctx, common.MapKeys(communityMap))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get communities: %v", err)
		return nil, errorx.Unknown
	}

	for i := range communities {
		communityMap[communities[i].ID] = &communities[i]
	}

	clientQuests := []model.Quest{}
	hiddenCount := 0
	for _, quest := range quests {
		if err := processValidationData(ctx, d.questFactory, false, &quest); err != nil {
			return nil, err
		}

		var category *entity.Category
		if quest.CategoryID.Valid {
			var ok bool
			category, ok = categoryMap[quest.CategoryID.String]
			if !ok {
				xcontext.Logger(ctx).Errorf("Invalid category id %s", quest.CategoryID.String)
				return nil, errorx.Unknown
			}
		}

		var community *entity.Community
		if quest.CommunityID.Valid {
			var ok bool
			community, ok = communityMap[quest.CommunityID.String]
			if !ok {
				xcontext.Logger(ctx).Errorf("Invalid community id %s", quest.CommunityID.String)
				return nil, errorx.Unknown
			}
		}

		q := convertQuest(&quest, convertCommunity(community, 0), convertCategory(category))
		if req.IncludeUnclaimableReason {
			reason, err := d.questFactory.IsClaimable(ctx, quest)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot determine not claimable reason: %v", err)
				return nil, errorx.Unknown
			}

			if reason != nil {
				if reason.Type == questclaim.UnclaimableByRecurrence {
					hiddenCount++
					continue // Hide this quest in case it is cannot claimable by recurrence.
				}

				q.UnclaimableReason = reason.Message
			}
		}

		clientQuests = append(clientQuests, q)
	}

	return &model.GetListQuestResponse{Quests: clientQuests, HiddenCount: hiddenCount}, nil
}

func (d *questDomain) GetTemplates(
	ctx context.Context, req *model.GetQuestTemplatesRequest,
) (*model.GetQuestTemplatestResponse, error) {
	// No need to bound the limit parameter because the number of quests is
	// usually small. Moreover, the frontend can get all quests to allow user
	// searching quests.
	// If the limit is not set, this method will return all quests by default.
	if req.Limit == 0 {
		req.Limit = -1
	}

	quests, err := d.questRepo.GetTemplates(ctx, repository.SearchQuestFilter{
		Q:      req.Q,
		Offset: req.Offset,
		Limit:  req.Limit,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get list of quest templates: %v", err)
		return nil, errorx.Unknown
	}

	categories, err := d.categoryRepo.GetTemplates(ctx)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get category: %v", err)
		return nil, errorx.Unknown
	}

	categoryMap := map[string]*entity.Category{}
	for i := range categories {
		categoryMap[categories[i].ID] = &categories[i]
	}

	clientQuests := []model.Quest{}
	for _, quest := range quests {
		var category *entity.Category
		if quest.CategoryID.Valid {
			var ok bool
			category, ok = categoryMap[quest.CategoryID.String]
			if !ok {
				xcontext.Logger(ctx).Errorf("Invalid category id %s", quest.CategoryID.String)
				return nil, errorx.Unknown
			}
		}

		clientQuests = append(clientQuests,
			convertQuest(&quest, model.Community{}, convertCategory(category)))
	}

	return &model.GetQuestTemplatestResponse{Templates: clientQuests}, nil
}

func (d *questDomain) ParseTemplate(
	ctx context.Context, req *model.ParseQuestTemplatesRequest,
) (*model.ParseQuestTemplatestResponse, error) {
	quest, err := d.questRepo.GetByID(ctx, req.TemplateID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get template: %v", err)
		return nil, errorx.Unknown
	}

	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	owner, err := d.userRepo.GetByID(ctx, community.CreatedBy)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get community owner: %v", err)
		return nil, errorx.Unknown
	}

	var category *entity.Category
	if quest.CategoryID.Valid {
		category, err = d.categoryRepo.GetByID(ctx, quest.CategoryID.String)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get category: %v", err)
			return nil, errorx.Unknown
		}
	}

	clientQuest := convertQuest(quest, model.Community{}, convertCategory(category))
	templateData := map[string]any{
		"owner": model.User{
			ID:            owner.ID,
			WalletAddress: owner.WalletAddress.String,
			Name:          owner.Name,
			Role:          string(owner.Role),
		},
		"community": model.Community{
			CreatedAt:    community.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:    community.UpdatedAt.Format(time.RFC3339Nano),
			CreatedBy:    community.CreatedBy,
			Introduction: string(community.Introduction),
			Handle:       community.Handle,
			DisplayName:  community.DisplayName,
			Twitter:      community.Twitter,
			Discord:      community.Discord,
		},
	}

	clientQuest.Title, err = common.ExecuteTemplate(clientQuest.Title, templateData)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot execute template of title: %v", err)
		return nil, errorx.Unknown
	}

	clientQuest.Description, err = common.ExecuteTemplate(clientQuest.Description, templateData)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot execute template of description: %v", err)
		return nil, errorx.Unknown
	}

	return &model.ParseQuestTemplatestResponse{Quest: clientQuest}, nil
}

func (d *questDomain) Update(
	ctx context.Context, req *model.UpdateQuestRequest,
) (*model.UpdateQuestResponse, error) {
	if req.ID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty id")
	}

	quest, err := d.questRepo.GetByID(ctx, req.ID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get quest: %v", err)
		return nil, errorx.Unknown
	}

	quest.Title = req.Title
	quest.Description = []byte(req.Description)
	quest.IsHighlight = req.IsHighlight

	if err = d.roleVerifier.Verify(ctx, quest.CommunityID.String, entity.AdminGroup...); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	quest.Status, err = enum.ToEnum[entity.QuestStatusType](req.Status)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid quest status: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid quest status %s", req.Status)
	}

	quest.Type, err = enum.ToEnum[entity.QuestType](req.Type)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid quest type: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid quest type %s", req.Type)
	}

	quest.Recurrence, err = enum.ToEnum[entity.RecurrenceType](req.Recurrence)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid recurrence: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid recurrence %s", req.Recurrence)
	}

	quest.ConditionOp, err = enum.ToEnum[entity.ConditionOpType](req.ConditionOp)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid condition op: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid condition op %s", req.ConditionOp)
	}

	for _, r := range req.Rewards {
		rType, err := enum.ToEnum[entity.RewardType](r.Type)
		if err != nil {
			return nil, errorx.New(errorx.BadRequest, "Invalid reward type %s", r.Type)
		}

		reward, err := d.questFactory.NewReward(ctx, *quest, rType, r.Data)
		if err != nil {
			return nil, err
		}

		quest.Rewards = append(quest.Rewards, entity.Reward{Type: rType, Data: structs.Map(reward)})
	}

	for _, c := range req.Conditions {
		ctype, err := enum.ToEnum[entity.ConditionType](c.Type)
		if err != nil {
			return nil, errorx.New(errorx.BadRequest, "Invalid condition type %s", c.Type)
		}

		condition, err := d.questFactory.NewCondition(ctx, ctype, c.Data)
		if err != nil {
			return nil, err
		}

		quest.Conditions = append(quest.Conditions, entity.Condition{Type: ctype, Data: structs.Map(condition)})
	}

	processor, err := d.questFactory.NewProcessor(ctx, *quest, req.ValidationData)
	if err != nil {
		return nil, err
	}
	quest.ValidationData = structs.Map(processor)

	if req.CategoryID != "" {
		quest.CategoryID = sql.NullString{Valid: true, String: req.CategoryID}
		category, err := d.categoryRepo.GetByID(ctx, req.CategoryID)
		if err != nil {
			xcontext.Logger(ctx).Debugf("Invalid category: %v", err)
			return nil, errorx.New(errorx.NotFound, "Invalid category")
		}

		if category.CommunityID.String != quest.CommunityID.String {
			return nil, errorx.New(errorx.BadRequest, "Category doesn't belong to community")
		}
	}

	changedPoints := int64(req.Points) - int64(quest.Points)
	quest.Points = req.Points

	err = d.questRepo.Update(ctx, quest)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update quest: %v", err)
		return nil, errorx.Unknown
	}

	if changedPoints != 0 && quest.CommunityID.Valid {
		followers, err := d.followerRepo.GetListByCommunityID(ctx, quest.CommunityID.String)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get list followers when changing point: %v", err)
			return nil, errorx.Unknown
		}

		for _, f := range followers {
			var err error
			if changedPoints > 0 {
				err = d.followerRepo.IncreasePoint(
					ctx, f.UserID, f.CommunityID, uint64(changedPoints), false)
			} else {
				// Currently, changedPoints is a negative number, DecreasePoint
				// receives a unsigned interger, so we must use the opposite
				// number of changedPoints.
				err = d.followerRepo.DecreasePoint(
					ctx, f.UserID, f.CommunityID, uint64(-changedPoints), false)
			}
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot change points of follower: %v", err)
				return nil, errorx.Unknown
			}
		}

		claimedQuests, err := d.claimedQuestRepo.GetList(
			ctx, &repository.ClaimedQuestFilter{
				CommunityID: quest.CommunityID.String,
				QuestIDs:    []string{quest.ID},
				Status:      []entity.ClaimedQuestStatus{entity.Accepted, entity.AutoAccepted},
			})
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get claimed quest of quests when changing point: %v", err)
			return nil, errorx.Unknown
		}

		for _, cq := range claimedQuests {
			err := d.leaderboard.ChangePointLeaderboard(
				ctx, int64(changedPoints), cq.ReviewedAt.Time, cq.UserID, quest.CommunityID.String)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot update leaderboard: %v", err)
				return nil, errorx.Unknown
			}
		}
	}

	return &model.UpdateQuestResponse{}, nil
}

func (d *questDomain) Delete(ctx context.Context, req *model.DeleteQuestRequest) (*model.DeleteQuestResponse, error) {
	if req.ID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty id")
	}

	quest, err := d.questRepo.GetByID(ctx, req.ID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get quest: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.roleVerifier.Verify(ctx, quest.CommunityID.String, entity.AdminGroup...); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	if err := d.questRepo.Delete(ctx, &entity.Quest{
		Base: entity.Base{ID: req.ID},
	}); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot delete quest: %v", err)
		return nil, errorx.Unknown
	}

	return &model.DeleteQuestResponse{}, nil
}
