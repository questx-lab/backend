package domain

import (
	"context"
	"database/sql"
	"time"

	"github.com/fatih/structs"
	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/questclaim"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/api/discord"
	"github.com/questx-lab/backend/pkg/api/telegram"
	"github.com/questx-lab/backend/pkg/api/twitter"
	"github.com/questx-lab/backend/pkg/enum"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
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
	roleVerifier     *common.CommunityRoleVerifier
	questFactory     questclaim.Factory
}

func NewQuestDomain(
	questRepo repository.QuestRepository,
	communityRepo repository.CommunityRepository,
	categoryRepo repository.CategoryRepository,
	collaboratorRepo repository.CollaboratorRepository,
	userRepo repository.UserRepository,
	claimedQuestRepo repository.ClaimedQuestRepository,
	oauth2Repo repository.OAuth2Repository,
	transactionRepo repository.TransactionRepository,
	twitterEndpoint twitter.IEndpoint,
	discordEndpoint discord.IEndpoint,
	telegramEndpoint telegram.IEndpoint,
) *questDomain {
	roleVerifier := common.NewCommunityRoleVerifier(collaboratorRepo, userRepo)

	return &questDomain{
		questRepo:        questRepo,
		communityRepo:    communityRepo,
		categoryRepo:     categoryRepo,
		claimedQuestRepo: claimedQuestRepo,
		userRepo:         userRepo,
		roleVerifier:     common.NewCommunityRoleVerifier(collaboratorRepo, userRepo),
		questFactory: questclaim.NewFactory(
			claimedQuestRepo,
			questRepo,
			communityRepo,
			nil, // No need to know follower information when creating quest.
			oauth2Repo,
			nil, // No need to know user aggregate when creating quest.
			userRepo,
			transactionRepo,
			roleVerifier,
			twitterEndpoint,
			discordEndpoint,
			telegramEndpoint,
		),
	}
}

func (d *questDomain) Create(
	ctx context.Context, req *model.CreateQuestRequest,
) (*model.CreateQuestResponse, error) {
	if err := d.roleVerifier.Verify(ctx, req.CommunityID, entity.AdminGroup...); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	quest := &entity.Quest{
		Base:        entity.Base{ID: uuid.NewString()},
		CommunityID: sql.NullString{Valid: true, String: req.CommunityID},
		IsTemplate:  false,
		Title:       req.Title,
		Description: []byte(req.Description),
		IsHighlight: req.IsHighlight,
	}

	if req.CommunityID == "" {
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
			return nil, errorx.New(errorx.BadRequest, "Invalid reward type %s", r.Type)
		}

		reward, err := d.questFactory.NewReward(ctx, *quest, rType, r.Data)
		if err != nil {
			xcontext.Logger(ctx).Debugf("Invalid reward data: %v", err)
			return nil, errorx.New(errorx.BadRequest, "Invalid reward data")
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
			xcontext.Logger(ctx).Debugf("Invalid condition data: %v", err)
			return nil, errorx.New(errorx.BadRequest, "Invalid condition data")
		}

		quest.Conditions = append(quest.Conditions, entity.Condition{Type: ctype, Data: structs.Map(condition)})
	}

	var processor questclaim.Processor
	if req.CommunityID != "" {
		processor, err = d.questFactory.NewProcessor(ctx, *quest, req.ValidationData)
	} else {
		processor, err = d.questFactory.LoadProcessor(ctx, *quest, req.ValidationData)
	}

	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid validation data: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid validation data")
	}
	quest.ValidationData = structs.Map(processor)

	if req.CategoryID != "" {
		quest.CategoryID = sql.NullString{Valid: true, String: req.CategoryID}
		category, err := d.categoryRepo.GetByID(ctx, req.CategoryID)
		if err != nil {
			xcontext.Logger(ctx).Debugf("Invalid category: %v", err)
			return nil, errorx.New(errorx.NotFound, "Invalid category")
		}

		if category.CommunityID.String != req.CommunityID {
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

	clientQuest := &model.GetQuestResponse{
		ID:             quest.ID,
		CommunityID:    quest.CommunityID.String,
		Type:           string(quest.Type),
		Status:         string(quest.Status),
		Title:          quest.Title,
		Description:    string(quest.Description),
		Recurrence:     string(quest.Recurrence),
		ValidationData: quest.ValidationData,
		Rewards:        rewardEntityToModel(quest.Rewards),
		ConditionOp:    string(quest.ConditionOp),
		Conditions:     conditionEntityToModel(quest.Conditions),
		CreatedAt:      quest.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:      quest.UpdatedAt.Format(time.RFC3339Nano),
	}

	var category *entity.Category
	if quest.CategoryID.Valid {
		category, err = d.categoryRepo.GetByID(ctx, quest.CategoryID.String)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get category: %v", err)
			return nil, errorx.Unknown
		}
	}

	if category != nil {
		clientQuest.Category = &model.Category{
			ID: category.ID, Name: category.Name,
		}
	}

	if req.IncludeUnclaimableReason {
		reason, err := d.questFactory.IsClaimable(ctx, *quest)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot determine not claimable reason: %v", err)
			return nil, errorx.Unknown
		}

		clientQuest.UnclaimableReason = reason
	}

	return clientQuest, nil
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

	quests, err := d.questRepo.GetList(ctx, repository.SearchQuestFilter{
		Q:           req.Q,
		CommunityID: req.CommunityID,
		CategoryID:  req.CategoryID,
		Offset:      req.Offset,
		Limit:       req.Limit,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get list of quests: %v", err)
		return nil, errorx.Unknown
	}

	categories, err := d.categoryRepo.GetList(ctx, req.CommunityID)
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
		q := model.Quest{
			ID:             quest.ID,
			CommunityID:    quest.CommunityID.String,
			Type:           string(quest.Type),
			Title:          quest.Title,
			Status:         string(quest.Status),
			Recurrence:     string(quest.Recurrence),
			Description:    string(quest.Description),
			ValidationData: quest.ValidationData,
			Rewards:        rewardEntityToModel(quest.Rewards),
			ConditionOp:    string(quest.ConditionOp),
			Conditions:     conditionEntityToModel(quest.Conditions),
			CreatedAt:      quest.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:      quest.UpdatedAt.Format(time.RFC3339Nano),
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

		if category != nil {
			q.Category = &model.Category{
				ID: category.ID, Name: category.Name,
			}
		}

		if req.IncludeUnclaimableReason {
			reason, err := d.questFactory.IsClaimable(ctx, quest)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot determine not claimable reason: %v", err)
				return nil, errorx.Unknown
			}

			q.UnclaimableReason = reason
		}

		clientQuests = append(clientQuests, q)
	}

	return &model.GetListQuestResponse{Quests: clientQuests}, nil
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

	categories, err := d.categoryRepo.GetList(ctx, "")
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
		template := model.Quest{
			ID:             quest.ID,
			CommunityID:    quest.CommunityID.String,
			Type:           string(quest.Type),
			Title:          quest.Title,
			Status:         string(quest.Status),
			Recurrence:     string(quest.Recurrence),
			Description:    string(quest.Description),
			ValidationData: quest.ValidationData,
			Rewards:        rewardEntityToModel(quest.Rewards),
			ConditionOp:    string(quest.ConditionOp),
			Conditions:     conditionEntityToModel(quest.Conditions),
			CreatedAt:      quest.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:      quest.UpdatedAt.Format(time.RFC3339Nano),
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

		if category != nil {
			template.Category = &model.Category{
				ID: category.ID, Name: category.Name,
			}
		}

		clientQuests = append(clientQuests, template)
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

	community, err := d.communityRepo.GetByID(ctx, req.CommunityID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	owner, err := d.userRepo.GetByID(ctx, community.CreatedBy)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get community owner: %v", err)
		return nil, errorx.Unknown
	}

	clientQuest := model.Quest{
		ID:             quest.ID,
		CommunityID:    quest.CommunityID.String,
		Type:           string(quest.Type),
		Title:          quest.Title,
		Status:         string(quest.Status),
		Recurrence:     string(quest.Recurrence),
		Description:    string(quest.Description),
		ValidationData: quest.ValidationData,
		Rewards:        rewardEntityToModel(quest.Rewards),
		ConditionOp:    string(quest.ConditionOp),
		Conditions:     conditionEntityToModel(quest.Conditions),
		CreatedAt:      quest.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:      quest.UpdatedAt.Format(time.RFC3339Nano),
	}

	var category *entity.Category
	if quest.CategoryID.Valid {
		category, err = d.categoryRepo.GetByID(ctx, quest.CategoryID.String)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get category: %v", err)
			return nil, errorx.Unknown
		}
	}

	if category != nil {
		clientQuest.Category = &model.Category{
			ID: category.ID, Name: category.Name,
		}
	}

	templateData := map[string]any{
		"owner": model.User{
			ID:      owner.ID,
			Address: owner.Address.String,
			Name:    owner.Name,
			Role:    string(owner.Role),
		},
		"community": model.Community{
			ID:           community.ID,
			CreatedAt:    community.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:    community.UpdatedAt.Format(time.RFC3339Nano),
			CreatedBy:    community.CreatedBy,
			Introduction: string(community.Introduction),
			Name:         community.Name,
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
			xcontext.Logger(ctx).Debugf("Invalid reward data: %v", err)
			return nil, errorx.New(errorx.BadRequest, "Invalid reward data")
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
			xcontext.Logger(ctx).Debugf("Invalid condition data: %v", err)
			return nil, errorx.New(errorx.BadRequest, "Invalid condition data")
		}

		quest.Conditions = append(quest.Conditions, entity.Condition{Type: ctype, Data: structs.Map(condition)})
	}

	processor, err := d.questFactory.NewProcessor(ctx, *quest, req.ValidationData)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid validation data: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid validation data")
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

	err = d.questRepo.Update(ctx, quest)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update quest: %v", err)
		return nil, errorx.Unknown
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
