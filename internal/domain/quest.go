package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/domain/questutil"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/enum"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/router"
)

type QuestDomain interface {
	Create(router.Context, *model.CreateQuestRequest) (*model.CreateQuestResponse, error)
	Get(router.Context, *model.GetQuestRequest) (*model.GetQuestResponse, error)
	GetList(router.Context, *model.GetListQuestRequest) (*model.GetListQuestResponse, error)
}

type questDomain struct {
	questRepo    repository.QuestRepository
	projectRepo  repository.ProjectRepository
	categoryRepo repository.CategoryRepository
}

func NewQuestDomain(
	questRepo repository.QuestRepository,
	projectRepo repository.ProjectRepository,
	catecategoryRepo repository.CategoryRepository,
) *questDomain {
	return &questDomain{
		questRepo:    questRepo,
		projectRepo:  projectRepo,
		categoryRepo: catecategoryRepo,
	}
}

func (d *questDomain) Create(
	ctx router.Context, req *model.CreateQuestRequest,
) (*model.CreateQuestResponse, error) {
	if req.ProjectID == "" {
		return nil, errorx.New(errorx.PermissionDenied, "Only admin can create quest template")
	}

	project, err := d.projectRepo.GetByID(ctx, req.ProjectID)
	if err != nil {
		ctx.Logger().Errorf("Cannot get project: %v", err)
		return nil, errorx.Unknown
	}

	if project.CreatedBy != ctx.GetUserID() {
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	questType, err := enum.ToEnum[entity.QuestType](req.Type)
	if err != nil {
		ctx.Logger().Debugf("Invalid quest type: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid quest type")
	}

	recurrence, err := enum.ToEnum[entity.RecurrenceType](req.Recurrence)
	if err != nil {
		ctx.Logger().Debugf("Invalid recurrence: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid recurrence")
	}

	conditionOp, err := enum.ToEnum[entity.ConditionOpType](req.ConditionOp)
	if err != nil {
		ctx.Logger().Debugf("Invalid condition op: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid condition op")
	}

	awards := []entity.Award{}
	awardFactory := questutil.NewAwardFactory()
	for _, a := range req.Awards {
		atype, err := enum.ToEnum[entity.AwardType](a.Type)
		if err != nil {
			return nil, errorx.New(errorx.BadRequest, "Invalid award type")
		}

		data := entity.Award{Type: atype, Value: a.Value}
		_, err = awardFactory.New(ctx, data)
		if err != nil {
			ctx.Logger().Debugf("Invalid award data: %v", err)
			return nil, errorx.New(errorx.BadRequest, "Invalid award data")
		}

		awards = append(awards, data)
	}

	conditions := []entity.Condition{}
	conditionFactory := questutil.NewConditionFactory(nil, d.questRepo)
	for _, c := range req.Conditions {
		ctype, err := enum.ToEnum[entity.ConditionType](c.Type)
		if err != nil {
			return nil, errorx.New(errorx.BadRequest, "Invalid condition type")
		}

		data := entity.Condition{Type: ctype, Op: c.Op, Value: c.Value}
		_, err = conditionFactory.New(ctx, data)
		if err != nil {
			ctx.Logger().Debugf("Invalid condition data: %v", err)
			return nil, errorx.New(errorx.BadRequest, "Invalid condition data")
		}

		conditions = append(conditions, data)
	}

	validatorFactory := questutil.NewValidatorFactory()
	if _, err := validatorFactory.New(ctx, questType, req.ValidationData); err != nil {
		ctx.Logger().Debugf("Invalid validation data: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid validation data")
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
		Status:         entity.Draft,
		ValidationData: req.ValidationData, // TODO: create a validator interface
		Awards:         awards,             // TODO: create award interface
		ConditionOp:    conditionOp,
		Conditions:     conditions, // TODO: create condition interface
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

func (d *questDomain) Get(ctx router.Context, req *model.GetQuestRequest) (*model.GetQuestResponse, error) {
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

	return &model.GetQuestResponse{
		ProjectID:      quest.ProjectID,
		Type:           string(quest.Type),
		Status:         string(quest.Status),
		Title:          quest.Title,
		Description:    quest.Description,
		Categories:     quest.CategoryIDs,
		Recurrence:     string(quest.Recurrence),
		ValidationData: quest.ValidationData,
		Awards:         awards,
		ConditionOp:    string(quest.ConditionOp),
		Conditions:     conditions,
		CreatedAt:      quest.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:      quest.UpdatedAt.Format(time.RFC3339Nano),
	}, nil
}

func (d *questDomain) GetList(
	ctx router.Context, req *model.GetListQuestRequest,
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
