package domain

import (
	"time"

	"github.com/fatih/structs"
	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/questclaim"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/api/discord"
	"github.com/questx-lab/backend/pkg/api/twitter"
	"github.com/questx-lab/backend/pkg/enum"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type QuestDomain interface {
	Create(xcontext.Context, *model.CreateQuestRequest) (*model.CreateQuestResponse, error)
	Update(xcontext.Context, *model.UpdateQuestRequest) (*model.UpdateQuestResponse, error)
	Get(xcontext.Context, *model.GetQuestRequest) (*model.GetQuestResponse, error)
	GetList(xcontext.Context, *model.GetListQuestRequest) (*model.GetListQuestResponse, error)
}

type questDomain struct {
	questRepo       repository.QuestRepository
	projectRepo     repository.ProjectRepository
	categoryRepo    repository.CategoryRepository
	roleVerifier    *common.ProjectRoleVerifier
	twitterEndpoint twitter.IEndpoint
	discordEndpoint discord.IEndpoint
	questFactory    questclaim.Factory
}

func NewQuestDomain(
	questRepo repository.QuestRepository,
	projectRepo repository.ProjectRepository,
	categoryRepo repository.CategoryRepository,
	collaboratorRepo repository.CollaboratorRepository,
	userRepo repository.UserRepository,
	twitterEndpoint twitter.IEndpoint,
	discordEndpoint discord.IEndpoint,
) *questDomain {
	return &questDomain{
		questRepo:       questRepo,
		projectRepo:     projectRepo,
		categoryRepo:    categoryRepo,
		roleVerifier:    common.NewProjectRoleVerifier(collaboratorRepo, userRepo),
		twitterEndpoint: twitterEndpoint,
		discordEndpoint: discordEndpoint,
		questFactory: questclaim.NewFactory(
			nil, // No need claimedQuestRepo when creating quest.
			questRepo,
			projectRepo,
			nil, // No need to know participant information when creating quest.
			nil, // No need to know user oauth2 id when creating quest.
			nil, // No need to know user aggregate when creating quest.
			twitterEndpoint,
			discordEndpoint,
		),
	}
}

func (d *questDomain) Create(
	ctx xcontext.Context, req *model.CreateQuestRequest,
) (*model.CreateQuestResponse, error) {
	if req.ProjectID == "" {
		return nil, errorx.New(errorx.PermissionDenied, "Only admin can create quest template")
	}

	if err := d.roleVerifier.Verify(ctx, req.ProjectID, entity.AdminGroup...); err != nil {
		ctx.Logger().Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	quest := &entity.Quest{
		Base:        entity.Base{ID: uuid.NewString()},
		ProjectID:   req.ProjectID,
		Title:       req.Title,
		Description: []byte(req.Description),
		Status:      entity.QuestDraft,
	}

	var err error
	quest.Type, err = enum.ToEnum[entity.QuestType](req.Type)
	if err != nil {
		ctx.Logger().Debugf("Invalid quest type: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid quest type %s", req.Type)
	}

	quest.Recurrence, err = enum.ToEnum[entity.RecurrenceType](req.Recurrence)
	if err != nil {
		ctx.Logger().Debugf("Invalid recurrence: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid recurrence %s", req.Recurrence)
	}

	quest.ConditionOp, err = enum.ToEnum[entity.ConditionOpType](req.ConditionOp)
	if err != nil {
		ctx.Logger().Debugf("Invalid condition op: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid condition op %s", req.ConditionOp)
	}

	for _, r := range req.Rewards {
		rType, err := enum.ToEnum[entity.RewardType](r.Type)
		if err != nil {
			return nil, errorx.New(errorx.BadRequest, "Invalid reward type %s", r.Type)
		}

		reward, err := d.questFactory.NewReward(ctx, *quest, rType, r.Data)
		if err != nil {
			ctx.Logger().Debugf("Invalid reward data: %v", err)
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
			ctx.Logger().Debugf("Invalid condition data: %v", err)
			return nil, errorx.New(errorx.BadRequest, "Invalid condition data")
		}

		quest.Conditions = append(quest.Conditions, entity.Condition{Type: ctype, Data: structs.Map(condition)})
	}

	processor, err := d.questFactory.NewProcessor(ctx, *quest, req.ValidationData)
	if err != nil {
		ctx.Logger().Debugf("Invalid validation data: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid validation data")
	}
	quest.ValidationData = structs.Map(processor)

	quest.CategoryIDs = req.Categories
	if err := d.categoryRepo.IsExisted(ctx, req.ProjectID, req.Categories...); err != nil {
		return nil, errorx.New(errorx.NotFound, "Invalid category")
	}

	err = d.questRepo.Create(ctx, quest)
	if err != nil {
		ctx.Logger().Errorf("Cannot create quest: %v", err)
		return nil, errorx.Unknown
	}

	return &model.CreateQuestResponse{
		ID: quest.ID,
	}, nil
}

func (d *questDomain) Get(ctx xcontext.Context, req *model.GetQuestRequest) (*model.GetQuestResponse, error) {
	if req.ID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty id")
	}

	quest, err := d.questRepo.GetByID(ctx, req.ID)
	if err != nil {
		ctx.Logger().Errorf("Cannot get quest: %v", err)
		return nil, errorx.Unknown
	}

	rewards := []model.Reward{}
	for _, a := range quest.Rewards {
		rewards = append(rewards, model.Reward{Type: string(a.Type), Data: a.Data})
	}

	conditions := []model.Condition{}
	for _, c := range quest.Conditions {
		conditions = append(conditions, model.Condition{Type: string(c.Type), Data: c.Data})
	}

	return &model.GetQuestResponse{
		ProjectID:      quest.ProjectID,
		Type:           string(quest.Type),
		Status:         string(quest.Status),
		Title:          quest.Title,
		Description:    string(quest.Description),
		Categories:     quest.CategoryIDs,
		Recurrence:     string(quest.Recurrence),
		ValidationData: quest.ValidationData,
		Rewards:        rewards,
		ConditionOp:    string(quest.ConditionOp),
		Conditions:     conditions,
		CreatedAt:      quest.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:      quest.UpdatedAt.Format(time.RFC3339Nano),
	}, nil
}

func (d *questDomain) GetList(
	ctx xcontext.Context, req *model.GetListQuestRequest,
) (*model.GetListQuestResponse, error) {
	if req.Limit == 0 {
		req.Limit = ctx.Configs().ApiServer.DefaultLimit
	}

	if req.Limit < 0 {
		return nil, errorx.New(errorx.BadRequest, "Limit must be positive")
	}

	if req.Limit > ctx.Configs().ApiServer.MaxLimit {
		return nil, errorx.New(errorx.BadRequest, "Exceed the maximum of limit")
	}

	quests, err := d.questRepo.GetList(ctx, req.ProjectID, req.Offset, req.Limit)
	if err != nil {
		ctx.Logger().Errorf("Cannot get list of quests: %v", err)
		return nil, errorx.Unknown
	}

	clientQuests := []model.Quest{}
	for _, quest := range quests {
		rewards := []model.Reward{}
		for _, r := range quest.Rewards {
			rewards = append(rewards, model.Reward{Type: string(r.Type), Data: r.Data})
		}

		conditions := []model.Condition{}
		for _, c := range quest.Conditions {
			conditions = append(conditions, model.Condition{Type: string(c.Type), Data: c.Data})
		}

		q := model.Quest{
			ID:             quest.ID,
			ProjectID:      quest.ProjectID,
			Type:           string(quest.Type),
			Title:          quest.Title,
			Status:         string(quest.Status),
			Recurrence:     string(quest.Recurrence),
			Categories:     quest.CategoryIDs,
			Description:    string(quest.Description),
			ValidationData: quest.ValidationData,
			Rewards:        rewards,
			ConditionOp:    string(quest.ConditionOp),
			Conditions:     conditions,
			CreatedAt:      quest.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:      quest.UpdatedAt.Format(time.RFC3339Nano),
		}

		clientQuests = append(clientQuests, q)
	}

	return &model.GetListQuestResponse{Quests: clientQuests}, nil
}

func (d *questDomain) Update(
	ctx xcontext.Context, req *model.UpdateQuestRequest,
) (*model.UpdateQuestResponse, error) {
	if req.ID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty id")
	}

	quest, err := d.questRepo.GetByID(ctx, req.ID)
	if err != nil {
		ctx.Logger().Errorf("Cannot get quest: %v", err)
		return nil, errorx.Unknown
	}

	if err = d.roleVerifier.Verify(ctx, quest.ProjectID, entity.AdminGroup...); err != nil {
		ctx.Logger().Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	status, err := enum.ToEnum[entity.QuestStatusType](req.Status)
	if err != nil {
		ctx.Logger().Debugf("Invalid quest status: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid quest status %s", req.Status)
	}

	err = d.questRepo.Update(ctx, &entity.Quest{
		Base:   entity.Base{ID: req.ID},
		Status: status,
	})
	if err != nil {
		ctx.Logger().Errorf("Cannot update quest: %v", err)
		return nil, errorx.Unknown
	}

	return &model.UpdateQuestResponse{}, nil
}
