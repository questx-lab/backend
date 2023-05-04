package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/questclaim"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/api/discord"
	"github.com/questx-lab/backend/pkg/api/twitter"
	"github.com/questx-lab/backend/pkg/crypto"
	"github.com/questx-lab/backend/pkg/dateutil"
	"github.com/questx-lab/backend/pkg/enum"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type ClaimedQuestDomain interface {
	Claim(xcontext.Context, *model.ClaimQuestRequest) (*model.ClaimQuestResponse, error)
	Get(xcontext.Context, *model.GetClaimedQuestRequest) (*model.GetClaimedQuestResponse, error)
	GetList(xcontext.Context, *model.GetListClaimedQuestRequest) (*model.GetListClaimedQuestResponse, error)
	ReviewClaimedQuest(xcontext.Context, *model.ReviewClaimedQuestRequest) (*model.ReviewClaimedQuestResponse, error)
	GiveReward(xcontext.Context, *model.GiveRewardRequest) (*model.GiveRewardResponse, error)
}

type claimedQuestDomain struct {
	claimedQuestRepo  repository.ClaimedQuestRepository
	questRepo         repository.QuestRepository
	participantRepo   repository.ParticipantRepository
	userAggregateRepo repository.UserAggregateRepository
	oauth2Repo        repository.OAuth2Repository
	projectRepo       repository.ProjectRepository
	roleVerifier      *common.ProjectRoleVerifier
	userRepo          repository.UserRepository
	twitterEndpoint   twitter.IEndpoint
	discordEndpoint   discord.IEndpoint
	questFactory      questclaim.Factory
}

func NewClaimedQuestDomain(
	claimedQuestRepo repository.ClaimedQuestRepository,
	questRepo repository.QuestRepository,
	collaboratorRepo repository.CollaboratorRepository,
	participantRepo repository.ParticipantRepository,
	oauth2Repo repository.OAuth2Repository,
	userAggregateRepo repository.UserAggregateRepository,
	userRepo repository.UserRepository,
	projectRepo repository.ProjectRepository,
	twitterEndpoint twitter.IEndpoint,
	discordEndpoint discord.IEndpoint,
) *claimedQuestDomain {
	questFactory := questclaim.NewFactory(
		claimedQuestRepo,
		questRepo,
		projectRepo,
		participantRepo,
		oauth2Repo,
		userAggregateRepo,
		twitterEndpoint,
		discordEndpoint,
	)

	return &claimedQuestDomain{
		claimedQuestRepo:  claimedQuestRepo,
		questRepo:         questRepo,
		participantRepo:   participantRepo,
		oauth2Repo:        oauth2Repo,
		userRepo:          userRepo,
		projectRepo:       projectRepo,
		roleVerifier:      common.NewProjectRoleVerifier(collaboratorRepo, userRepo),
		userAggregateRepo: userAggregateRepo,
		twitterEndpoint:   twitterEndpoint,
		discordEndpoint:   discordEndpoint,
		questFactory:      questFactory,
	}
}

func (d *claimedQuestDomain) Claim(
	ctx xcontext.Context, req *model.ClaimQuestRequest,
) (*model.ClaimQuestResponse, error) {
	quest, err := d.questRepo.GetByID(ctx, req.QuestID)
	if err != nil {
		ctx.Logger().Errorf("Cannot get quest: %v", err)
		return nil, errorx.Unknown
	}

	if quest.Status != entity.QuestActive {
		return nil, errorx.New(errorx.Unavailable, "Only allow to claim active quests")
	}

	// Check if user follows the project.
	requestUserID := xcontext.GetRequestUserID(ctx)
	_, err = d.participantRepo.Get(ctx, requestUserID, quest.ProjectID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.Logger().Errorf("Cannot get the participant: %v", err)
			return nil, errorx.Unknown
		}

		// If the user has not followed project yet, he will follow it automatically.
		err = d.participantRepo.Create(ctx, &entity.Participant{
			UserID:     requestUserID,
			ProjectID:  quest.ProjectID,
			InviteCode: crypto.GenerateRandomAlphabet(9),
		})
		if err != nil {
			ctx.Logger().Errorf("Cannot auto follow the project: %v", err)
			return nil, errorx.Unknown
		}
	}

	if err != nil {
		ctx.Logger().Errorf("Cannot create quest factory of user: %v", err)
		return nil, errorx.Unknown
	}

	// Check the condition and recurrence.
	claimable, err := d.isClaimable(ctx, *quest)
	if err != nil {
		ctx.Logger().Errorf("Cannot determine claimable: %v", err)
		return nil, errorx.Unknown
	}

	if !claimable {
		return nil, errorx.New(errorx.Unavailable, "This quest cannot be claimed now")
	}

	// Auto review the action/input of user with validation data. After this step, we can
	// determine if the quest user claimed is accepted, rejected, or need a manual review.
	processor, err := d.questFactory.LoadProcessor(ctx, *quest, quest.ValidationData)
	if err != nil {
		ctx.Logger().Debugf("Invalid validation data: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid validation data")
	}

	// Get the last claimed quest
	userID := xcontext.GetRequestUserID(ctx)
	lastClaimedQuest, err := d.claimedQuestRepo.GetLast(ctx, userID, quest.ID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.Logger().Errorf("Cannot get claimed quest: %v", err)
			return nil, errorx.Unknown
		}
	}

	result, err := processor.GetActionForClaim(ctx, lastClaimedQuest, req.Input)
	if err != nil {
		return nil, err
	}

	var status entity.ClaimedQuestStatus
	switch result {
	case questclaim.Accepted:
		status = entity.AutoAccepted
	case questclaim.Rejected:
		status = entity.AutoRejected
	case questclaim.NeedManualReview:
		status = entity.Pending
	}

	// Store the ClaimedQuest into database.
	claimedQuest := &entity.ClaimedQuest{
		Base:    entity.Base{ID: uuid.NewString()},
		QuestID: req.QuestID,
		UserID:  xcontext.GetRequestUserID(ctx),
		Status:  status,
		Input:   req.Input,
	}

	if status != entity.Pending {
		claimedQuest.ReviewerAt = time.Now()
	}

	// GiveReward can write something to database.
	ctx.BeginTx()
	defer ctx.RollbackTx()

	err = d.claimedQuestRepo.Create(ctx, claimedQuest)
	if err != nil {
		ctx.Logger().Errorf("Cannot claim quest: %v", err)
		return nil, errorx.Unknown
	}

	// Give reward to user if the claimed quest is accepted.
	if status == entity.AutoAccepted {
		for _, data := range quest.Rewards {
			reward, err := d.questFactory.LoadReward(ctx, *quest, data.Type, data.Data)
			if err != nil {
				ctx.Logger().Errorf("Invalid reward data: %v", err)
				return nil, errorx.Unknown
			}

			if err := reward.Give(ctx, xcontext.GetRequestUserID(ctx)); err != nil {
				return nil, err
			}
		}

		if err := d.increaseTask(ctx, quest.ProjectID, claimedQuest.UserID); err != nil {
			ctx.Logger().Errorf("Unable to increase number of task: %v", err)
			return nil, errorx.New(errorx.Internal, "Unable to increase number of task")
		}
	}

	ctx.CommitTx()
	return &model.ClaimQuestResponse{ID: claimedQuest.ID, Status: string(status)}, nil
}

func (d *claimedQuestDomain) Get(
	ctx xcontext.Context, req *model.GetClaimedQuestRequest,
) (*model.GetClaimedQuestResponse, error) {
	if req.ID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty id")
	}

	claimedQuest, err := d.claimedQuestRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found claimed quest")
		}

		ctx.Logger().Errorf("Cannot get claimed quest: %v", err)
		return nil, errorx.Unknown
	}

	quest, err := d.questRepo.GetByID(ctx, claimedQuest.QuestID)
	if err != nil {
		ctx.Logger().Errorf("Cannot get quest: %v", err)
		return nil, errorx.Unknown
	}

	if err = d.roleVerifier.Verify(ctx, quest.ProjectID, entity.AdminGroup...); err != nil {
		ctx.Logger().Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	return &model.GetClaimedQuestResponse{
		QuestID:    claimedQuest.QuestID,
		UserID:     claimedQuest.UserID,
		Input:      claimedQuest.Input,
		Status:     string(claimedQuest.Status),
		ReviewerID: claimedQuest.ReviewerID,
		ReviewerAt: claimedQuest.ReviewerAt.Format(time.RFC3339Nano),
		Comment:    claimedQuest.Comment,
	}, nil
}

func (d *claimedQuestDomain) GetList(
	ctx xcontext.Context, req *model.GetListClaimedQuestRequest,
) (*model.GetListClaimedQuestResponse, error) {
	if req.ProjectID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty project id")
	}

	if err := d.roleVerifier.Verify(ctx, req.ProjectID, entity.ReviewGroup...); err != nil {
		ctx.Logger().Errorf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	if req.Limit == 0 {
		req.Limit = ctx.Configs().ApiServer.DefaultLimit
	}

	if req.Limit < 0 {
		return nil, errorx.New(errorx.BadRequest, "Limit must be positive")
	}

	if req.Limit > ctx.Configs().ApiServer.MaxLimit {
		return nil, errorx.New(errorx.BadRequest, "Exceed the maximum of limit")
	}

	var statusFilter []entity.ClaimedQuestStatus
	if req.FilterStatus != "" {
		statuses := strings.Split(req.FilterStatus, ",")
		for _, status := range statuses {
			statusEnum, err := enum.ToEnum[entity.ClaimedQuestStatus](status)
			if err != nil {
				ctx.Logger().Debugf("Invalid claimed quest status: %v", err)
				return nil, errorx.New(errorx.BadRequest, "Invalid status filter")
			}

			statusFilter = append(statusFilter, statusEnum)
		}
	}

	result, err := d.claimedQuestRepo.GetList(
		ctx,
		&repository.ClaimedQuestFilter{
			ProjectID: req.ProjectID,
			Status:    statusFilter,
			QuestID:   req.FilterQuestID,
			UserID:    req.FilterUserID,
		},
		req.Offset,
		req.Limit,
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found any claimed quest")
		}

		ctx.Logger().Errorf("Cannot get list claimed quest: %v", err)
		return nil, errorx.Unknown
	}

	claimedQuests := []model.ClaimedQuest{}
	for _, q := range result {
		claimedQuests = append(claimedQuests, model.ClaimedQuest{
			QuestID:    q.QuestID,
			UserID:     q.UserID,
			Status:     string(q.Status),
			ReviewerID: q.ReviewerID,
			ReviewerAt: q.ReviewerAt.Format(time.RFC3339Nano),
		})
	}

	return &model.GetListClaimedQuestResponse{ClaimedQuests: claimedQuests}, nil
}

func (d *claimedQuestDomain) isClaimable(ctx xcontext.Context, quest entity.Quest) (bool, error) {
	// Check conditions.
	finalCondition := true
	if quest.ConditionOp == entity.Or && len(quest.Conditions) > 0 {
		finalCondition = false
	}

	for _, c := range quest.Conditions {
		condition, err := d.questFactory.LoadCondition(ctx, c.Type, c.Data)
		if err != nil {
			return false, err
		}

		b, err := condition.Check(ctx)
		if err != nil {
			return false, err
		}

		if quest.ConditionOp == entity.And {
			finalCondition = finalCondition && b
		} else {
			finalCondition = finalCondition || b
		}
	}

	if !finalCondition {
		return false, nil
	}

	// Check recurrence.
	requestUserID := xcontext.GetRequestUserID(ctx)
	lastClaimedQuest, err := d.claimedQuestRepo.GetLastPendingOrAccepted(ctx, requestUserID, quest.ID)
	if err != nil {
		// The user has not claimed this quest yet.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil
		}
		return false, err
	}

	// If the user claimed the quest before, this quest cannot be claimed again until the next
	// recurrence.
	switch quest.Recurrence {
	case entity.Once:
		return false, nil

	case entity.Daily:
		return lastClaimedQuest.CreatedAt.Day() != time.Now().Day(), nil

	case entity.Weekly:
		_, lastWeek := lastClaimedQuest.CreatedAt.ISOWeek()
		_, currentWeek := time.Now().ISOWeek()
		return lastWeek != currentWeek, nil

	case entity.Monthly:
		return lastClaimedQuest.CreatedAt.Month() != time.Now().Month(), nil

	default:
		return false, fmt.Errorf("invalid recurrence %s", quest.Recurrence)
	}
}

func (d *claimedQuestDomain) ReviewClaimedQuest(ctx xcontext.Context, req *model.ReviewClaimedQuestRequest) (*model.ReviewClaimedQuestResponse, error) {
	if req.ID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty id")
	}

	reviewAction, err := enum.ToEnum[entity.ClaimedQuestStatus](req.Action)
	if err != nil {
		ctx.Logger().Debugf("Invalid review action: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid action")
	}

	if reviewAction != entity.Accepted && reviewAction != entity.Rejected {
		return nil, errorx.New(errorx.BadRequest, "Action must be accepted or rejected")
	}

	claimedQuest, err := d.claimedQuestRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found claimed quest")
		}

		ctx.Logger().Errorf("Cannot get claimed quest: %v", err)
		return nil, errorx.Unknown
	}

	if claimedQuest.Status != entity.Pending {
		return nil, errorx.New(errorx.BadRequest, "Claimed quest must be pending")
	}

	quest, err := d.questRepo.GetByID(ctx, claimedQuest.QuestID)
	if err != nil {
		ctx.Logger().Errorf("Cannot get quest: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.roleVerifier.Verify(ctx, quest.ProjectID, entity.ReviewGroup...); err != nil {
		ctx.Logger().Errorf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	ctx.BeginTx()
	defer ctx.RollbackTx()

	requestUserID := xcontext.GetRequestUserID(ctx)
	if err := d.claimedQuestRepo.UpdateReviewByID(ctx, req.ID, &entity.ClaimedQuest{
		Status:     reviewAction,
		ReviewerID: requestUserID,
		ReviewerAt: time.Now(),
	}); err != nil {
		ctx.Logger().Errorf("Unable to update status: %v", err)
		return nil, errorx.New(errorx.Internal, "Unable to approve this claim quest")
	}

	if err != nil {
		ctx.Logger().Errorf("Cannot create quest factory of user: %v", err)
		return nil, errorx.Unknown
	}

	for _, data := range quest.Rewards {
		reward, err := d.questFactory.LoadReward(ctx, *quest, data.Type, data.Data)
		if err != nil {
			ctx.Logger().Errorf("Invalid reward data: %v", err)
			return nil, errorx.Unknown
		}

		if err := reward.Give(ctx, claimedQuest.UserID); err != nil {
			return nil, err
		}
	}

	if err := d.increaseTask(ctx, quest.ProjectID, claimedQuest.UserID); err != nil {
		ctx.Logger().Errorf("Unable to increase number of task: %v", err)
		return nil, errorx.New(errorx.Internal, "Unable to increase number of task")
	}

	ctx.CommitTx()
	return &model.ReviewClaimedQuestResponse{}, nil
}

func (d *claimedQuestDomain) increaseTask(ctx xcontext.Context, projectID, userID string) error {
	for _, r := range entity.UserAggregateRangeList {
		rangeValue, err := dateutil.GetCurrentValueByRange(r)
		if err != nil {
			return err
		}

		if err := d.userAggregateRepo.Upsert(ctx, &entity.UserAggregate{
			ProjectID:  projectID,
			UserID:     userID,
			Range:      r,
			RangeValue: rangeValue,
			TotalTask:  1,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (d *claimedQuestDomain) GiveReward(
	ctx xcontext.Context, req *model.GiveRewardRequest,
) (*model.GiveRewardResponse, error) {
	if err := d.roleVerifier.Verify(ctx, req.ProjectID, entity.Owner); err != nil {
		ctx.Logger().Debugf("Permission denined when give reward: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Only project owner can give reward directly")
	}

	_, err := d.participantRepo.Get(ctx, req.UserID, req.ProjectID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.Logger().Errorf("Cannot get the participant: %v", err)
			return nil, errorx.Unknown
		}

		return nil, errorx.New(errorx.Unavailable, "User must follow the project before")
	}

	rewardType, err := enum.ToEnum[entity.RewardType](req.Type)
	if err != nil {
		ctx.Logger().Debugf("Invalid reward type: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid reward type %s", req.Type)
	}

	if err != nil {
		ctx.Logger().Errorf("Cannot create quest factory of user: %v", err)
		return nil, errorx.Unknown
	}

	// Create a fake quest of this project.
	fakeQuest := entity.Quest{ProjectID: req.ProjectID}
	reward, err := d.questFactory.NewReward(ctx, fakeQuest, rewardType, req.Data)
	if err != nil {
		ctx.Logger().Debugf("Invalid reward: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid reward")
	}

	if err := reward.Give(ctx, req.UserID); err != nil {
		ctx.Logger().Warnf("Cannot give reward to user: %v", err)
		return nil, errorx.New(errorx.Unavailable, "Cannot give reward to user")
	}

	return &model.GiveRewardResponse{}, nil
}
