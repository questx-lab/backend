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
		return nil, errorx.NewGeneric(nil, "permission denied")
	}

	project, err := d.projectRepo.GetByID(ctx, req.ProjectID)
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

func (d *questDomain) Get(ctx router.Context, req *model.GetQuestRequest) (*model.GetQuestResponse, error) {
	if req.ID == "" {
		return nil, errorx.NewGeneric(nil, "not allow empty id")
	}

	quest, err := d.questRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, errorx.NewGeneric(err, "cannot get quest")
	}

	awards := []model.Award{}
	err = json.Unmarshal([]byte(quest.Awards), &awards)
	if err != nil {
		return nil, errorx.NewGeneric(err, "unable to execute the quest")
	}

	conditions := []model.Condition{}
	err = json.Unmarshal([]byte(quest.Conditions), &conditions)
	if err != nil {
		return nil, errorx.NewGeneric(err, "unable to execute the quest")
	}

	return &model.GetQuestResponse{
		ProjectID:      quest.ProjectID,
		Type:           enum.ToString(quest.Type),
		Status:         enum.ToString(quest.Status),
		Title:          quest.Title,
		Description:    quest.Description,
		Categories:     strings.Split(quest.CategoryIDs, ","),
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
		return nil, errorx.NewGeneric(nil, "limit must be positive")
	}

	if req.Limit > 50 {
		return nil, errorx.NewGeneric(nil, "exceed the maximum of limit")
	}

	quests, err := d.questRepo.GetListShortForm(ctx, req.ProjectID, req.Offset, req.Limit)
	if err != nil {
		return nil, errorx.NewGeneric(err, "cannot get quest")
	}

	shortQuests := []model.ShortQuest{}
	for _, quest := range quests {
		q := model.ShortQuest{
			ID:         quest.ID,
			Type:       enum.ToString(quest.Type),
			Title:      quest.Title,
			Status:     enum.ToString(quest.Status),
			Recurrence: enum.ToString(quest.Recurrence),
		}

		if quest.CategoryIDs != "" {
			q.Categories = strings.Split(quest.CategoryIDs, ",")
		}

		shortQuests = append(shortQuests, q)
	}

	return &model.GetListQuestResponse{
		Quests: shortQuests,
	}, nil
}
