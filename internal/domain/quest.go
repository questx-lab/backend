package domain

import (
	"encoding/json"
	"strings"

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
	GetShortForm(router.Context, *model.GetShortQuestRequest) (*model.GetShortQuestResponse, error)
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
		return nil, errorx.NewGeneric(nil, "permission denied")
	}

	project, err := d.projectRepo.GeyByID(ctx, req.ProjectID)
	if err != nil {
		return nil, errorx.NewGeneric(err, "cannot get the project with id %s", req.ProjectID)
	}

	if project.CreatedBy != ctx.GetUserID() {
		return nil, errorx.NewGeneric(nil, "permission denied")
	}

	questType, err := enum.ToEnum[entity.QuestType](req.Type)
	if err != nil {
		return nil, errorx.NewGeneric(err, "invalid quest type")
	}

	recurrence, err := enum.ToEnum[entity.QuestRecurrenceType](req.Recurrence)
	if err != nil {
		return nil, errorx.NewGeneric(err, "invalid recurrence")
	}

	conditionOp, err := enum.ToEnum[entity.QuestConditionOpType](req.ConditionOp)
	if err != nil {
		return nil, errorx.NewGeneric(err, "invalid condition operator")
	}

	awards, err := json.Marshal(req.Awards)
	if err != nil {
		return nil, errorx.NewGeneric(err, "invalid awards")
	}

	conditions, err := json.Marshal(req.Conditions)
	if err != nil {
		return nil, errorx.NewGeneric(err, "invalid conditions")
	}

	quest := &entity.Quest{
		Base:           entity.Base{ID: uuid.NewString()},
		ProjectID:      req.ProjectID,
		Title:          req.Title,
		Description:    req.Description,
		Type:           questType,
		CategoryIDs:    strings.Join(req.Categories, ","), // TODO: check after create category table
		Recurrence:     recurrence,
		Status:         entity.QuestStatusDraft,
		ValidationData: req.ValidationData, // TODO: create a validator interface
		Awards:         string(awards),     // TODO: create award interface
		ConditionOp:    conditionOp,
		Conditions:     string(conditions), // TODO: create condition interface
	}

	err = d.questRepo.Create(ctx, quest)
	if err != nil {
		return nil, errorx.NewGeneric(err, "cannot create quest")
	}

	return &model.CreateQuestResponse{
		ID: quest.ID,
	}, nil
}

func (d *questDomain) GetShortForm(
	ctx router.Context, req *model.GetShortQuestRequest,
) (*model.GetShortQuestResponse, error) {
	if req.ID == "" {
		return nil, errorx.NewGeneric(nil, "not allow empty id")
	}

	quest, err := d.questRepo.GetShortForm(ctx, req.ID)
	if err != nil {
		return nil, errorx.NewGeneric(err, "cannot get quest")
	}

	return &model.GetShortQuestResponse{
		ProjectID:  quest.ProjectID,
		Type:       enum.ToString(quest.Type),
		Title:      quest.Title,
		Categories: strings.Split(quest.CategoryIDs, ","),
		Recurrence: enum.ToString(quest.Recurrence),
	}, nil
}
