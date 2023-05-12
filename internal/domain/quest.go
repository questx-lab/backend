package domain

import (
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
	Create(xcontext.Context, *model.CreateQuestRequest) (*model.CreateQuestResponse, error)
	Update(xcontext.Context, *model.UpdateQuestRequest) (*model.UpdateQuestResponse, error)
	Get(xcontext.Context, *model.GetQuestRequest) (*model.GetQuestResponse, error)
	GetList(xcontext.Context, *model.GetListQuestRequest) (*model.GetListQuestResponse, error)
	Delete(xcontext.Context, *model.DeleteQuestRequest) (*model.DeleteQuestResponse, error)
	GetTemplates(xcontext.Context, *model.GetQuestTemplatesRequest) (*model.GetQuestTemplatestResponse, error)
	ParseTemplate(xcontext.Context, *model.ParseQuestTemplatesRequest) (*model.ParseQuestTemplatestResponse, error)
}

type questDomain struct {
	questRepo        repository.QuestRepository
	projectRepo      repository.ProjectRepository
	categoryRepo     repository.CategoryRepository
	claimedQuestRepo repository.ClaimedQuestRepository
	userRepo         repository.UserRepository
	roleVerifier     *common.ProjectRoleVerifier
	questFactory     questclaim.Factory
}

func NewQuestDomain(
	questRepo repository.QuestRepository,
	projectRepo repository.ProjectRepository,
	categoryRepo repository.CategoryRepository,
	collaboratorRepo repository.CollaboratorRepository,
	userRepo repository.UserRepository,
	claimedQuestRepo repository.ClaimedQuestRepository,
	oauth2Repo repository.OAuth2Repository,
	twitterEndpoint twitter.IEndpoint,
	discordEndpoint discord.IEndpoint,
	telegramEndpoint telegram.IEndpoint,
) *questDomain {
	roleVerifier := common.NewProjectRoleVerifier(collaboratorRepo, userRepo)

	return &questDomain{
		questRepo:        questRepo,
		projectRepo:      projectRepo,
		categoryRepo:     categoryRepo,
		claimedQuestRepo: claimedQuestRepo,
		userRepo:         userRepo,
		roleVerifier:     common.NewProjectRoleVerifier(collaboratorRepo, userRepo),
		questFactory: questclaim.NewFactory(
			claimedQuestRepo,
			questRepo,
			projectRepo,
			nil, // No need to know participant information when creating quest.
			oauth2Repo,
			nil, // No need to know user aggregate when creating quest.
			roleVerifier,
			twitterEndpoint,
			discordEndpoint,
			telegramEndpoint,
		),
	}
}

func (d *questDomain) Create(
	ctx xcontext.Context, req *model.CreateQuestRequest,
) (*model.CreateQuestResponse, error) {
	if err := d.roleVerifier.Verify(ctx, req.ProjectID, entity.AdminGroup...); err != nil {
		ctx.Logger().Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	quest := &entity.Quest{
		Base:        entity.Base{ID: uuid.NewString()},
		ProjectID:   sql.NullString{Valid: true, String: req.ProjectID},
		Title:       req.Title,
		Description: []byte(req.Description),
	}

	if req.ProjectID == "" {
		quest.ProjectID = sql.NullString{Valid: false}
	}

	var err error
	quest.Type, err = enum.ToEnum[entity.QuestType](req.Type)
	if err != nil {
		ctx.Logger().Debugf("Invalid quest type: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid quest type %s", req.Type)
	}

	quest.Status, err = enum.ToEnum[entity.QuestStatusType](req.Status)
	if err != nil {
		ctx.Logger().Debugf("Invalid quest status: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid quest status %s", req.Status)
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

	if err != nil {
		ctx.Logger().Errorf("Cannot create quest factory with user: %v", err)
		return nil, errorx.Unknown
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
		ctx.Logger().Debugf("Invalid category: %v", err)
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

	clientQuest := &model.GetQuestResponse{
		ID:             quest.ID,
		ProjectID:      quest.ProjectID.String,
		Type:           string(quest.Type),
		Status:         string(quest.Status),
		Title:          quest.Title,
		Description:    string(quest.Description),
		Categories:     quest.CategoryIDs,
		Recurrence:     string(quest.Recurrence),
		ValidationData: quest.ValidationData,
		Rewards:        rewardEntityToModel(quest.Rewards),
		ConditionOp:    string(quest.ConditionOp),
		Conditions:     conditionEntityToModel(quest.Conditions),
		CreatedAt:      quest.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:      quest.UpdatedAt.Format(time.RFC3339Nano),
	}

	if req.IncludeUnclaimableReason {
		reason, err := questclaim.IsClaimable(ctx, d.questFactory, d.claimedQuestRepo, *quest)
		if err != nil {
			ctx.Logger().Errorf("Cannot determine not claimable reason: %v", err)
			return nil, errorx.Unknown
		}

		clientQuest.UnclaimableReason = reason
	}

	return clientQuest, nil
}

func (d *questDomain) GetList(
	ctx xcontext.Context, req *model.GetListQuestRequest,
) (*model.GetListQuestResponse, error) {
	// No need to bound the limit parameter because the number of quests is
	// usually small. Moreover, the frontend can get all quests to allow user
	// searching quests.

	// If the limit is not set, this method will return all quests by default.
	if req.Limit == 0 {
		req.Limit = -1
	}

	quests, err := d.questRepo.GetList(ctx, repository.SearchQuestFilter{
		Q:         req.Q,
		ProjectID: req.ProjectID,
		Offset:    req.Offset,
		Limit:     req.Limit,
	})
	if err != nil {
		ctx.Logger().Errorf("Cannot get list of quests: %v", err)
		return nil, errorx.Unknown
	}

	clientQuests := []model.Quest{}
	for _, quest := range quests {
		q := model.Quest{
			ID:             quest.ID,
			ProjectID:      quest.ProjectID.String,
			Type:           string(quest.Type),
			Title:          quest.Title,
			Status:         string(quest.Status),
			Recurrence:     string(quest.Recurrence),
			Categories:     quest.CategoryIDs,
			Description:    string(quest.Description),
			ValidationData: quest.ValidationData,
			Rewards:        rewardEntityToModel(quest.Rewards),
			ConditionOp:    string(quest.ConditionOp),
			Conditions:     conditionEntityToModel(quest.Conditions),
			CreatedAt:      quest.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:      quest.UpdatedAt.Format(time.RFC3339Nano),
		}

		if req.IncludeUnclaimableReason {
			reason, err := questclaim.IsClaimable(ctx, d.questFactory, d.claimedQuestRepo, quest)
			if err != nil {
				ctx.Logger().Errorf("Cannot determine not claimable reason: %v", err)
				return nil, errorx.Unknown
			}

			q.UnclaimableReason = reason
		}

		clientQuests = append(clientQuests, q)
	}

	return &model.GetListQuestResponse{Quests: clientQuests}, nil
}

func (d *questDomain) GetTemplates(
	ctx xcontext.Context, req *model.GetQuestTemplatesRequest,
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
		ctx.Logger().Errorf("Cannot get list of quest templates: %v", err)
		return nil, errorx.Unknown
	}

	clientQuests := []model.Quest{}
	for _, quest := range quests {
		clientQuests = append(clientQuests, model.Quest{
			ID:             quest.ID,
			ProjectID:      quest.ProjectID.String,
			Type:           string(quest.Type),
			Title:          quest.Title,
			Status:         string(quest.Status),
			Recurrence:     string(quest.Recurrence),
			Categories:     quest.CategoryIDs,
			Description:    string(quest.Description),
			ValidationData: quest.ValidationData,
			Rewards:        rewardEntityToModel(quest.Rewards),
			ConditionOp:    string(quest.ConditionOp),
			Conditions:     conditionEntityToModel(quest.Conditions),
			CreatedAt:      quest.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:      quest.UpdatedAt.Format(time.RFC3339Nano),
		})
	}

	return &model.GetQuestTemplatestResponse{Quests: clientQuests}, nil
}

func (d *questDomain) ParseTemplate(
	ctx xcontext.Context, req *model.ParseQuestTemplatesRequest,
) (*model.ParseQuestTemplatestResponse, error) {
	quest, err := d.questRepo.GetByID(ctx, req.TemplateID)
	if err != nil {
		ctx.Logger().Errorf("Cannot get template: %v", err)
		return nil, errorx.Unknown
	}

	project, err := d.projectRepo.GetByID(ctx, req.ProjectID)
	if err != nil {
		ctx.Logger().Errorf("Cannot get project: %v", err)
		return nil, errorx.Unknown
	}

	owner, err := d.userRepo.GetByID(ctx, project.CreatedBy)
	if err != nil {
		ctx.Logger().Errorf("Cannot get project owner: %v", err)
		return nil, errorx.Unknown
	}

	clientQuest := model.Quest{
		ID:             quest.ID,
		ProjectID:      quest.ProjectID.String,
		Type:           string(quest.Type),
		Title:          quest.Title,
		Status:         string(quest.Status),
		Recurrence:     string(quest.Recurrence),
		Categories:     quest.CategoryIDs,
		Description:    string(quest.Description),
		ValidationData: quest.ValidationData,
		Rewards:        rewardEntityToModel(quest.Rewards),
		ConditionOp:    string(quest.ConditionOp),
		Conditions:     conditionEntityToModel(quest.Conditions),
		CreatedAt:      quest.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:      quest.UpdatedAt.Format(time.RFC3339Nano),
	}

	templateData := map[string]any{
		"owner": model.User{
			ID:      owner.ID,
			Address: owner.Address.String,
			Name:    owner.Name,
			Role:    string(owner.Role),
		},
		"project": model.Project{
			ID:           project.ID,
			CreatedAt:    project.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:    project.UpdatedAt.Format(time.RFC3339Nano),
			CreatedBy:    project.CreatedBy,
			Introduction: string(project.Introduction),
			Name:         project.Name,
			Twitter:      project.Twitter,
			Discord:      project.Discord,
		},
	}

	clientQuest.Title, err = common.ExecuteTemplate(clientQuest.Title, templateData)
	if err != nil {
		ctx.Logger().Errorf("Cannot execute template of title: %v", err)
		return nil, errorx.Unknown
	}

	clientQuest.Description, err = common.ExecuteTemplate(clientQuest.Description, templateData)
	if err != nil {
		ctx.Logger().Errorf("Cannot execute template of description: %v", err)
		return nil, errorx.Unknown
	}

	return &model.ParseQuestTemplatestResponse{Quest: clientQuest}, nil
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

	if err = d.roleVerifier.Verify(ctx, quest.ProjectID.String, entity.AdminGroup...); err != nil {
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
	if err := d.categoryRepo.IsExisted(ctx, quest.ProjectID.String, req.Categories...); err != nil {
		ctx.Logger().Debugf("Invalid category: %v", err)
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

	if err := d.roleVerifier.Verify(ctx, quest.ProjectID.String, entity.AdminGroup...); err != nil {
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
