package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/questclaim"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
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
	Delete(xcontext.Context, *model.DeleteQuestRequest) (*model.DeleteQuestResponse, error)
}

type questDomain struct {
	questRepo       repository.QuestRepository
	projectRepo     repository.ProjectRepository
	categoryRepo    repository.CategoryRepository
	roleVerifier    *common.ProjectRoleVerifier
	twitterEndpoint twitter.IEndpoint
}

func NewQuestDomain(
	questRepo repository.QuestRepository,
	projectRepo repository.ProjectRepository,
	categoryRepo repository.CategoryRepository,
	collaboratorRepo repository.CollaboratorRepository,
	twitterEndpoint twitter.IEndpoint,
) *questDomain {
	return &questDomain{
		questRepo:       questRepo,
		projectRepo:     projectRepo,
		categoryRepo:    categoryRepo,
		roleVerifier:    common.NewProjectRoleVerifier(collaboratorRepo),
		twitterEndpoint: twitterEndpoint,
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

	questType, err := enum.ToEnum[entity.QuestType](req.Type)
	if err != nil {
		ctx.Logger().Debugf("Invalid quest type: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid quest type %s", req.Type)
	}

	recurrence, err := enum.ToEnum[entity.RecurrenceType](req.Recurrence)
	if err != nil {
		ctx.Logger().Debugf("Invalid recurrence: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid recurrence %s", req.Recurrence)
	}

	conditionOp, err := enum.ToEnum[entity.ConditionOpType](req.ConditionOp)
	if err != nil {
		ctx.Logger().Debugf("Invalid condition op: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid condition op %s", req.ConditionOp)
	}

	awards := []entity.Award{}
	for _, a := range req.Awards {
		atype, err := enum.ToEnum[entity.AwardType](a.Type)
		if err != nil {
			return nil, errorx.New(errorx.BadRequest, "Invalid award type %s", a.Type)
		}

		data := entity.Award{Type: atype, Value: a.Value}
		_, err = questclaim.NewAward(ctx, nil, data)
		if err != nil {
			ctx.Logger().Debugf("Invalid award data: %v", err)
			return nil, errorx.New(errorx.BadRequest, "Invalid award data")
		}

		awards = append(awards, data)
	}

	conditions := []entity.Condition{}
	for _, c := range req.Conditions {
		ctype, err := enum.ToEnum[entity.ConditionType](c.Type)
		if err != nil {
			return nil, errorx.New(errorx.BadRequest, "Invalid condition type %s", c.Type)
		}

		data := entity.Condition{Type: ctype, Op: c.Op, Value: c.Value}
		_, err = questclaim.NewCondition(ctx, nil, d.questRepo, data)
		if err != nil {
			ctx.Logger().Debugf("Invalid condition data: %v", err)
			return nil, errorx.New(errorx.BadRequest, "Invalid condition data")
		}

		conditions = append(conditions, data)
	}

	processor, err := questclaim.NewProcessor(ctx, d.twitterEndpoint, questType, req.ValidationData)
	if err != nil {
		ctx.Logger().Debugf("Invalid validation data: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid validation data")
	}

	validationData, err := json.Marshal(processor)
	if err != nil {
		ctx.Logger().Debugf("Cannot marshal validation data: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Cannot marshal validation data")
	}

	if err := d.categoryRepo.IsExisted(ctx, req.ProjectID, req.Categories...); err != nil {
		return nil, errorx.New(errorx.NotFound, "Invalid category")
	}

	quest := &entity.Quest{
		Base:           entity.Base{ID: uuid.NewString()},
		ProjectID:      req.ProjectID,
		Title:          req.Title,
		Description:    req.Description,
		Type:           questType,
		CategoryIDs:    req.Categories,
		Recurrence:     recurrence,
		Status:         entity.QuestDraft,
		ValidationData: string(validationData),
		Awards:         awards,
		ConditionOp:    conditionOp,
		Conditions:     conditions,
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

	awards := []model.Award{}
	for _, a := range quest.Awards {
		awards = append(awards, model.Award{Type: string(a.Type), Value: a.Value})
	}

	conditions := []model.Condition{}
	for _, c := range quest.Conditions {
		conditions = append(conditions, model.Condition{Type: string(c.Type), Op: c.Op, Value: c.Value})
	}

	validationData := map[string]any{}
	err = json.Unmarshal([]byte(quest.ValidationData), &validationData)
	if err != nil {
		ctx.Logger().Errorf("Cannot unmarshal validation data: %v", err)
		return nil, errorx.Unknown
	}

	return &model.GetQuestResponse{
		ProjectID:      quest.ProjectID,
		Type:           string(quest.Type),
		Status:         string(quest.Status),
		Title:          quest.Title,
		Description:    quest.Description,
		Categories:     quest.CategoryIDs,
		Recurrence:     string(quest.Recurrence),
		ValidationData: validationData,
		Awards:         awards,
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
		req.Limit = 1
	}

	if req.Limit < 0 {
		return nil, errorx.New(errorx.BadRequest, "Limit must be positive")
	}

	if req.Limit > 50 {
		return nil, errorx.New(errorx.BadRequest, "Exceed the maximum of limit")
	}

	quests, err := d.questRepo.GetList(ctx, req.ProjectID, req.Offset, req.Limit)
	if err != nil {
		ctx.Logger().Errorf("Cannot get list of quests: %v", err)
		return nil, errorx.Unknown
	}

	shortQuests := []model.ShortQuest{}
	for _, quest := range quests {
		q := model.ShortQuest{
			ID:         quest.ID,
			Type:       string(quest.Type),
			Title:      quest.Title,
			Status:     string(quest.Status),
			Recurrence: string(quest.Recurrence),
			Categories: quest.CategoryIDs,
		}

		shortQuests = append(shortQuests, q)
	}

	return &model.GetListQuestResponse{
		Quests: shortQuests,
	}, nil
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

	quest.Status, err = enum.ToEnum[entity.QuestStatusType](req.Status)
	if err != nil {
		ctx.Logger().Debugf("Invalid quest status: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid quest status %s", req.Status)
	}

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
	if err := d.categoryRepo.IsExisted(ctx, quest.ProjectID, req.Categories...); err != nil {
		return nil, errorx.New(errorx.NotFound, "Invalid category")
	}

	err = d.questRepo.Update(ctx, quest)
	if err != nil {
		ctx.Logger().Errorf("Cannot update quest: %v", err)
		return nil, errorx.Unknown
	}

	return &model.UpdateQuestResponse{}, nil
}

func (d *questDomain) Delete(ctx xcontext.Context, req *model.DeleteQuestRequest) (*model.DeleteQuestResponse, error) {
	if req.ID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty id")
	}

	quest, err := d.questRepo.GetByID(ctx, req.ID)
	if err != nil {
		ctx.Logger().Errorf("Cannot get quest: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.roleVerifier.Verify(ctx, quest.ProjectID, entity.AdminGroup...); err != nil {
		ctx.Logger().Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	if err := d.questRepo.Delete(ctx, &entity.Quest{
		Base: entity.Base{ID: req.ID},
	}); err != nil {
		ctx.Logger().Errorf("Cannot delete quest: %v", err)
		return nil, errorx.Unknown
	}

	return &model.DeleteQuestResponse{}, nil
}
