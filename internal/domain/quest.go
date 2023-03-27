package domain

import (
	"github.com/google/uuid"
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
	questRepo   repository.QuestRepository
	projectRepo repository.ProjectRepository
}

func NewQuestDomain(
	questRepo repository.QuestRepository,
	projectRepo repository.ProjectRepository,
) *questDomain {
	return &questDomain{questRepo: questRepo, projectRepo: projectRepo}
}

func (d *questDomain) Create(
	ctx router.Context, req *model.CreateQuestRequest,
) (*model.CreateQuestResponse, error) {
	if req.ProjectID == "" {
		// Only admin can create quest template.
		return nil, errorx.NewGeneric(nil, "Permission denied")
	}

	project, err := d.projectRepo.GetByID(ctx, req.ProjectID)
	if err != nil {
		return nil, errorx.NewGeneric(err, "Cannot get the project with id %s", req.ProjectID)
	}

	if project.CreatedBy != ctx.GetUserID() {
		return nil, errorx.NewGeneric(nil, "Permission denied")
	}

	questType, err := enum.ToEnum[entity.QuestType](req.Type)
	if err != nil {
		return nil, errorx.NewGeneric(err, "Invalid quest type")
	}

	recurrence, err := enum.ToEnum[entity.QuestRecurrenceType](req.Recurrence)
	if err != nil {
		return nil, errorx.NewGeneric(err, "Invalid recurrence")
	}

	conditionOp, err := enum.ToEnum[entity.QuestConditionOpType](req.ConditionOp)
	if err != nil {
		return nil, errorx.NewGeneric(err, "Invalid condition operator")
	}

	awards := []entity.Award{}
	for _, a := range req.Awards {
		awards = append(awards, entity.Award{Type: a.Type, Value: a.Value})
	}

	conditions := []entity.Condition{}
	for _, c := range req.Conditions {
		conditions = append(conditions, entity.Condition{Type: c.Type, Op: c.Op, Value: c.Value})
	}

	quest := &entity.Quest{
		Base:           entity.Base{ID: uuid.NewString()},
		ProjectID:      req.ProjectID,
		Title:          req.Title,
		Description:    req.Description,
		Type:           questType,
		CategoryIDs:    req.Categories, // TODO: check after create category table
		Recurrence:     recurrence,
		Status:         entity.QuestStatusDraft,
		ValidationData: req.ValidationData, // TODO: create a validator interface
		Awards:         awards,             // TODO: create award interface
		ConditionOp:    conditionOp,
		Conditions:     conditions, // TODO: create condition interface
	}

	err = d.questRepo.Create(ctx, quest)
	if err != nil {
		return nil, errorx.NewGeneric(err, "Cannot create quest")
	}

	return &model.CreateQuestResponse{
		ID: quest.ID,
	}, nil
}

func (d *questDomain) Get(ctx router.Context, req *model.GetQuestRequest) (*model.GetQuestResponse, error) {
	if req.ID == "" {
		return nil, errorx.NewGeneric(nil, "Not allow empty id")
	}

	quest, err := d.questRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, errorx.NewGeneric(err, "Cannot get quest")
	}

	awards := []model.Award{}
	for _, a := range quest.Awards {
		awards = append(awards, model.Award{Type: a.Type, Value: a.Value})
	}

	conditions := []model.Condition{}
	for _, c := range quest.Conditions {
		conditions = append(conditions, model.Condition{Type: c.Type, Op: c.Op, Value: c.Value})
	}

	return &model.GetQuestResponse{
		ProjectID:      quest.ProjectID,
		Type:           enum.ToString(quest.Type),
		Status:         enum.ToString(quest.Status),
		Title:          quest.Title,
		Description:    quest.Description,
		Categories:     quest.CategoryIDs,
		Recurrence:     enum.ToString(quest.Recurrence),
		ValidationData: quest.ValidationData,
		Awards:         awards,
		ConditionOp:    enum.ToString(quest.ConditionOp),
		Conditions:     conditions,
	}, nil
}

func (d *questDomain) GetList(
	ctx router.Context, req *model.GetListQuestRequest,
) (*model.GetListQuestResponse, error) {
	// If the limit is not set, the default value is 1.
	if req.Limit == 0 {
		req.Limit = 1
	}

	if req.Limit < 0 {
		return nil, errorx.NewGeneric(nil, "Limit must be positive")
	}

	if req.Limit > 50 {
		return nil, errorx.NewGeneric(nil, "Exceed the maximum of limit")
	}

	quests, err := d.questRepo.GetListShortForm(ctx, req.ProjectID, req.Offset, req.Limit)
	if err != nil {
		return nil, errorx.NewGeneric(err, "Cannot get quest")
	}

	shortQuests := []model.ShortQuest{}
	for _, quest := range quests {
		q := model.ShortQuest{
			ID:         quest.ID,
			Type:       enum.ToString(quest.Type),
			Title:      quest.Title,
			Status:     enum.ToString(quest.Status),
			Recurrence: enum.ToString(quest.Recurrence),
			Categories: quest.CategoryIDs,
		}

		shortQuests = append(shortQuests, q)
	}

	return &model.GetListQuestResponse{
		Quests: shortQuests,
	}, nil
}
