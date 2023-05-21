package domain

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/badge"
	"github.com/questx-lab/backend/internal/domain/questclaim"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/api/discord"
	"github.com/questx-lab/backend/pkg/api/telegram"
	"github.com/questx-lab/backend/pkg/api/twitter"
	"github.com/questx-lab/backend/pkg/dateutil"
	"github.com/questx-lab/backend/pkg/enum"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type ClaimedQuestDomain interface {
	Claim(context.Context, *model.ClaimQuestRequest) (*model.ClaimQuestResponse, error)
	ClaimReferral(context.Context, *model.ClaimReferralRequest) (*model.ClaimReferralResponse, error)
	Get(context.Context, *model.GetClaimedQuestRequest) (*model.GetClaimedQuestResponse, error)
	GetList(context.Context, *model.GetListClaimedQuestRequest) (*model.GetListClaimedQuestResponse, error)
	Review(context.Context, *model.ReviewRequest) (*model.ReviewResponse, error)
	ReviewAll(context.Context, *model.ReviewAllRequest) (*model.ReviewAllResponse, error)
	GiveReward(context.Context, *model.GiveRewardRequest) (*model.GiveRewardResponse, error)
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
	badgeManager      *badge.Manager
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
	transactionRepo repository.TransactionRepository,
	twitterEndpoint twitter.IEndpoint,
	discordEndpoint discord.IEndpoint,
	telegramEndpoint telegram.IEndpoint,
	badgeManager *badge.Manager,
) *claimedQuestDomain {
	roleVerifier := common.NewProjectRoleVerifier(collaboratorRepo, userRepo)

	questFactory := questclaim.NewFactory(
		claimedQuestRepo,
		questRepo,
		projectRepo,
		participantRepo,
		oauth2Repo,
		userAggregateRepo,
		userRepo,
		transactionRepo,
		roleVerifier,
		twitterEndpoint,
		discordEndpoint,
		telegramEndpoint,
	)

	return &claimedQuestDomain{
		claimedQuestRepo:  claimedQuestRepo,
		questRepo:         questRepo,
		participantRepo:   participantRepo,
		oauth2Repo:        oauth2Repo,
		userRepo:          userRepo,
		projectRepo:       projectRepo,
		roleVerifier:      roleVerifier,
		userAggregateRepo: userAggregateRepo,
		twitterEndpoint:   twitterEndpoint,
		discordEndpoint:   discordEndpoint,
		questFactory:      questFactory,
		badgeManager:      badgeManager,
	}
}

func (d *claimedQuestDomain) Claim(
	ctx context.Context, req *model.ClaimQuestRequest,
) (*model.ClaimQuestResponse, error) {
	quest, err := d.questRepo.GetByID(ctx, req.QuestID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get quest: %v", err)
		return nil, errorx.Unknown
	}

	if quest.Status != entity.QuestActive {
		return nil, errorx.New(errorx.Unavailable, "Only allow to claim active quests")
	}

	if !quest.ProjectID.Valid {
		return nil, errorx.New(errorx.Unavailable, "Unable to claim a template")
	}

	// Check if user follows the project.
	requestUserID := xcontext.RequestUserID(ctx)
	_, err = d.participantRepo.Get(ctx, requestUserID, quest.ProjectID.String)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			xcontext.Logger(ctx).Errorf("Cannot get the participant: %v", err)
			return nil, errorx.Unknown
		}

		err := followProject(
			ctx,
			d.userRepo,
			d.projectRepo,
			d.participantRepo,
			nil,
			requestUserID, quest.ProjectID.String, "",
		)
		if err != nil {
			return nil, err
		}
	}

	// Check the condition and recurrence.
	reason, err := d.questFactory.IsClaimable(ctx, *quest)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot determine claimable: %v", err)
		return nil, errorx.Unknown
	}

	if reason != "" {
		return nil, errorx.New(errorx.Unavailable, reason)
	}

	// Auto review the action/input of user with validation data. After this step, we can
	// determine if the quest user claimed is accepted, rejected, or need a manual review.
	processor, err := d.questFactory.LoadProcessor(ctx, *quest, quest.ValidationData)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid validation data: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid validation data")
	}

	result, err := processor.GetActionForClaim(ctx, req.Input)
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
		UserID:  xcontext.RequestUserID(ctx),
		Status:  status,
		Input:   req.Input,
	}

	if status != entity.Pending {
		claimedQuest.ReviewedAt = time.Now()
	}

	// GiveReward can write something to database.
	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	// If the claimed quest is auto accepted or pending (even rejected after
	// reviewing), the streak will be stacked.
	if status != entity.AutoRejected {
		// Get the last claimed quest (accepted or pending of any quest) to
		// calculate streak.
		lastClaimedAnyQuest, err := d.claimedQuestRepo.GetLast(ctx, repository.GetLastClaimedQuestFilter{
			UserID: xcontext.RequestUserID(ctx), ProjectID: quest.ProjectID.String,
			Status: []entity.ClaimedQuestStatus{entity.Pending, entity.Accepted, entity.AutoAccepted},
		})
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			xcontext.Logger(ctx).Errorf("Cannot get claimed quest: %v", err)
			return nil, errorx.Unknown
		}

		if err == nil && dateutil.IsYesterday(lastClaimedAnyQuest.CreatedAt, claimedQuest.CreatedAt) {
			err := d.participantRepo.IncreaseStat(ctx, requestUserID, quest.ProjectID.String, 0, 1)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot increase streak: %v", err)
				return nil, errorx.Unknown
			}
		} else {
			// In this case, we cannot find the last claimed quest or the last
			// claimed time is not yesterday, we need to reset the streak.
			err := d.participantRepo.IncreaseStat(ctx, requestUserID, quest.ProjectID.String, 0, -1)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot increase streak: %v", err)
				return nil, errorx.Unknown
			}
		}

		err = d.badgeManager.
			WithBadges(badge.RainBowBadgeName).
			ScanAndGive(ctx, requestUserID, quest.ProjectID.String)
		if err != nil {
			return nil, err
		}
	}

	err = d.claimedQuestRepo.Create(ctx, claimedQuest)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot claim quest: %v", err)
		return nil, errorx.Unknown
	}

	// Give reward to user if the claimed quest is accepted.
	if status == entity.AutoAccepted {
		for _, data := range quest.Rewards {
			reward, err := d.questFactory.LoadReward(ctx, *quest, data.Type, data.Data)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Invalid reward data: %v", err)
				return nil, errorx.Unknown
			}

			if err := reward.Give(ctx, xcontext.RequestUserID(ctx), claimedQuest.ID); err != nil {
				return nil, err
			}
		}

		if err := d.increaseTask(ctx, quest.ProjectID.String, claimedQuest.UserID); err != nil {
			xcontext.Logger(ctx).Errorf("Unable to increase number of task: %v", err)
			return nil, errorx.New(errorx.Internal, "Unable to increase number of task")
		}

		err := d.badgeManager.
			WithBadges(badge.QuestWarriorBadgeName).
			ScanAndGive(ctx, requestUserID, quest.ProjectID.String)
		if err != nil {
			return nil, err
		}
	}

	xcontext.WithCommitDBTransaction(ctx)
	return &model.ClaimQuestResponse{ID: claimedQuest.ID, Status: string(status)}, nil
}

func (d *claimedQuestDomain) ClaimReferral(
	ctx context.Context, req *model.ClaimReferralRequest,
) (*model.ClaimReferralResponse, error) {
	requestUserID := xcontext.RequestUserID(ctx)
	projects, err := d.projectRepo.GetList(ctx, repository.GetListProjectFilter{
		ReferredBy:     requestUserID,
		ReferralStatus: entity.ReferralClaimable,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get claimable referral projects: %v", err)
		return nil, errorx.Unknown
	}

	if len(projects) == 0 {
		return nil, errorx.New(errorx.Unavailable, "Not found any claimable referral project")
	}

	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	allNames := []string{}
	projectIDs := []string{}
	for _, p := range projects {
		allNames = append(allNames, p.Name)
		projectIDs = append(projectIDs, p.ID)
	}

	coinReward, err := d.questFactory.NewReward(
		ctx,
		entity.Quest{},
		entity.CointReward,
		map[string]any{
			"note":       fmt.Sprintf("Referral reward of %s", strings.Join(allNames, " | ")),
			"token":      xcontext.Configs(ctx).Quest.InviteProjectRewardToken,
			"amount":     xcontext.Configs(ctx).Quest.InviteProjectRewardAmount * float64(len(projects)),
			"to_address": req.Address,
		},
	)
	if err != nil {
		return nil, errorx.Unknown
	}

	if err := coinReward.Give(ctx, requestUserID, ""); err != nil {
		return nil, err
	}

	err = d.projectRepo.UpdateReferralStatusByIDs(ctx, projectIDs, entity.ReferralClaimed)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update referral status of project to claimed: %v", err)
		return nil, errorx.Unknown
	}

	xcontext.WithCommitDBTransaction(ctx)
	return &model.ClaimReferralResponse{}, nil
}

func (d *claimedQuestDomain) Get(
	ctx context.Context, req *model.GetClaimedQuestRequest,
) (*model.GetClaimedQuestResponse, error) {
	if req.ID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty id")
	}

	claimedQuest, err := d.claimedQuestRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found claimed quest")
		}

		xcontext.Logger(ctx).Errorf("Cannot get claimed quest: %v", err)
		return nil, errorx.Unknown
	}

	quest, err := d.questRepo.GetByID(ctx, claimedQuest.QuestID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get quest: %v", err)
		return nil, errorx.Unknown
	}

	if err = d.roleVerifier.Verify(ctx, quest.ProjectID.String, entity.AdminGroup...); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	user, err := d.userRepo.GetByID(ctx, claimedQuest.UserID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get users: %v", err)
		return nil, errorx.Unknown
	}

	return &model.GetClaimedQuestResponse{
		ID:      claimedQuest.ID,
		QuestID: claimedQuest.QuestID,
		Quest: model.Quest{
			ID:             quest.ID,
			ProjectID:      quest.ProjectID.String,
			Type:           string(quest.Type),
			Status:         string(quest.Status),
			Title:          quest.Title,
			Description:    string(quest.Description),
			CategoryID:     quest.CategoryID.String,
			Recurrence:     string(quest.Recurrence),
			ValidationData: quest.ValidationData,
			Rewards:        rewardEntityToModel(quest.Rewards),
			ConditionOp:    string(quest.ConditionOp),
			Conditions:     conditionEntityToModel(quest.Conditions),
			CreatedAt:      quest.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:      quest.UpdatedAt.Format(time.RFC3339Nano),
		},
		UserID: claimedQuest.UserID,
		User: model.User{
			ID:      user.ID,
			Address: user.Address.String,
			Name:    user.Name,
			Role:    string(user.Role),
		},
		Input:      claimedQuest.Input,
		Status:     string(claimedQuest.Status),
		ReviewerID: claimedQuest.ReviewerID,
		ReviewedAt: claimedQuest.ReviewedAt.Format(time.RFC3339Nano),
		Comment:    claimedQuest.Comment,
	}, nil
}

func (d *claimedQuestDomain) GetList(
	ctx context.Context, req *model.GetListClaimedQuestRequest,
) (*model.GetListClaimedQuestResponse, error) {
	if req.ProjectID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty project id")
	}

	if err := d.roleVerifier.Verify(ctx, req.ProjectID, entity.ReviewGroup...); err != nil {
		xcontext.Logger(ctx).Errorf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	if req.Limit == 0 {
		req.Limit = xcontext.Configs(ctx).ApiServer.DefaultLimit
	}

	if req.Limit < 0 {
		return nil, errorx.New(errorx.BadRequest, "Limit must be positive")
	}

	if req.Limit > xcontext.Configs(ctx).ApiServer.MaxLimit {
		return nil, errorx.New(errorx.BadRequest, "Exceed the maximum of limit")
	}

	var statusFilter []entity.ClaimedQuestStatus
	if req.Status != "" {
		status := strings.Split(req.Status, ",")
		for _, s := range status {
			statusEnum, err := enum.ToEnum[entity.ClaimedQuestStatus](s)
			if err != nil {
				xcontext.Logger(ctx).Debugf("Invalid claimed quest status: %v", err)
				return nil, errorx.New(errorx.BadRequest, "Invalid status filter")
			}

			statusFilter = append(statusFilter, statusEnum)
		}
	}

	var recurrenceFilter []entity.RecurrenceType
	if req.Recurrence != "" {
		recurrences := strings.Split(req.Recurrence, ",")
		for _, recurrence := range recurrences {
			recurrenceEnum, err := enum.ToEnum[entity.RecurrenceType](recurrence)
			if err != nil {
				xcontext.Logger(ctx).Debugf("Invalid recurrence: %v", err)
				return nil, errorx.New(errorx.BadRequest, "Invalid recurrence filter")
			}

			recurrenceFilter = append(recurrenceFilter, recurrenceEnum)
		}
	}

	var userIDFilter []string
	if len(req.UserID) > 0 {
		userIDFilter = strings.Split(req.UserID, ",")
	}

	var questIDFilter []string
	if len(req.QuestID) > 0 {
		questIDFilter = strings.Split(req.QuestID, ",")
	}

	result, err := d.claimedQuestRepo.GetList(
		ctx,
		req.ProjectID,
		&repository.ClaimedQuestFilter{
			Status:      statusFilter,
			Recurrences: recurrenceFilter,
			QuestIDs:    questIDFilter,
			UserIDs:     userIDFilter,
			Offset:      req.Offset,
			Limit:       req.Limit,
		},
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found any claimed quest")
		}

		xcontext.Logger(ctx).Errorf("Cannot get list claimed quest: %v", err)
		return nil, errorx.Unknown
	}

	claimedQuests := []model.ClaimedQuest{}
	questSet := map[string]any{}
	userSet := map[string]any{}
	for _, cq := range result {
		claimedQuests = append(claimedQuests, model.ClaimedQuest{
			ID:         cq.ID,
			QuestID:    cq.QuestID,
			UserID:     cq.UserID,
			Status:     string(cq.Status),
			Comment:    cq.Comment,
			ReviewerID: cq.ReviewerID,
			ReviewedAt: cq.ReviewedAt.Format(time.RFC3339Nano),
		})

		questSet[cq.QuestID] = nil
		userSet[cq.UserID] = nil
	}

	quests, err := d.questRepo.GetByIDs(ctx, common.MapKeys(questSet))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get quests: %v", err)
		return nil, errorx.Unknown
	}

	questsInverse := map[string]entity.Quest{}
	for _, q := range quests {
		questsInverse[q.ID] = q
	}

	users, err := d.userRepo.GetByIDs(ctx, common.MapKeys(userSet))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get users: %v", err)
		return nil, errorx.Unknown
	}

	usersInverse := map[string]entity.User{}
	for _, u := range users {
		usersInverse[u.ID] = u
	}

	for i, cq := range claimedQuests {
		quest, ok := questsInverse[cq.QuestID]
		if !ok {
			xcontext.Logger(ctx).Errorf("Not found quest %s in claimed quest %s", cq.QuestID, cq.ID)
			return nil, errorx.Unknown
		}

		user, ok := usersInverse[cq.UserID]
		if !ok {
			xcontext.Logger(ctx).Errorf("Not found user %s in claimed quest %s", cq.UserID, cq.ID)
			return nil, errorx.Unknown
		}

		claimedQuests[i].Quest = model.Quest{
			ID:             quest.ID,
			ProjectID:      quest.ProjectID.String,
			Type:           string(quest.Type),
			Status:         string(quest.Status),
			Title:          quest.Title,
			Description:    string(quest.Description),
			CategoryID:     quest.CategoryID.String,
			Recurrence:     string(quest.Recurrence),
			ValidationData: quest.ValidationData,
			Rewards:        rewardEntityToModel(quest.Rewards),
			ConditionOp:    string(quest.ConditionOp),
			Conditions:     conditionEntityToModel(quest.Conditions),
			CreatedAt:      quest.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:      quest.UpdatedAt.Format(time.RFC3339Nano),
		}

		claimedQuests[i].User = model.User{
			ID:      user.ID,
			Address: user.Address.String,
			Name:    user.Name,
			Role:    string(user.Role),
		}
	}

	return &model.GetListClaimedQuestResponse{ClaimedQuests: claimedQuests}, nil
}

func (d *claimedQuestDomain) Review(
	ctx context.Context, req *model.ReviewRequest,
) (*model.ReviewResponse, error) {
	if len(req.IDs) == 0 {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty id")
	}

	reviewAction, err := enum.ToEnum[entity.ClaimedQuestStatus](req.Action)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid review action: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid action")
	}

	if reviewAction != entity.Accepted && reviewAction != entity.Rejected {
		return nil, errorx.New(errorx.BadRequest, "Action must be accepted or rejected")
	}

	firstClaimedQuest, err := d.claimedQuestRepo.GetByID(ctx, req.IDs[0])
	if err != nil {
		xcontext.Logger(ctx).Debugf("Cannot get the first claimed quest: %v", err)
		return nil, errorx.Unknown
	}

	firstQuest, err := d.questRepo.GetByID(ctx, firstClaimedQuest.QuestID)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Cannot get the first quest: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.roleVerifier.Verify(ctx, firstQuest.ProjectID.String, entity.ReviewGroup...); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	claimedQuests, err := d.claimedQuestRepo.GetByIDs(ctx, req.IDs)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get claimed quest by ids: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.review(ctx, claimedQuests, reviewAction, req.Comment); err != nil {
		return nil, err
	}

	return &model.ReviewResponse{}, nil
}

func (d *claimedQuestDomain) ReviewAll(
	ctx context.Context, req *model.ReviewAllRequest,
) (*model.ReviewAllResponse, error) {
	if req.ProjectID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow an empty project id")
	}

	if err := d.roleVerifier.Verify(ctx, req.ProjectID, entity.ReviewGroup...); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	reviewAction, err := enum.ToEnum[entity.ClaimedQuestStatus](req.Action)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid review action: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid action")
	}

	if reviewAction != entity.Accepted && reviewAction != entity.Rejected {
		return nil, errorx.New(errorx.BadRequest, "Action must be accepted or rejected")
	}

	var recurrenceFilter []entity.RecurrenceType
	for _, recurrence := range req.Recurrences {
		recurrenceEnum, err := enum.ToEnum[entity.RecurrenceType](recurrence)
		if err != nil {
			xcontext.Logger(ctx).Debugf("Invalid recurrence: %v", err)
			return nil, errorx.New(errorx.BadRequest, "Invalid recurrence filter")
		}

		recurrenceFilter = append(recurrenceFilter, recurrenceEnum)
	}

	claimedQuests, err := d.claimedQuestRepo.GetList(
		ctx,
		req.ProjectID,
		&repository.ClaimedQuestFilter{
			QuestIDs:    req.QuestIDs,
			UserIDs:     req.UserIDs,
			Status:      []entity.ClaimedQuestStatus{entity.Pending},
			Recurrences: recurrenceFilter,
			Offset:      0,
			Limit:       -1,
		},
	)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get claimed quest: %v", err)
		return nil, errorx.Unknown
	}

	excludeMap := map[string]any{}
	for _, id := range req.Excludes {
		excludeMap[id] = nil
	}

	finalClaimedQuests := []entity.ClaimedQuest{}
	for _, cq := range claimedQuests {
		if _, ok := excludeMap[cq.ID]; !ok {
			finalClaimedQuests = append(finalClaimedQuests, cq)
		}
	}

	if err := d.review(ctx, finalClaimedQuests, reviewAction, req.Comment); err != nil {
		return nil, err
	}

	return &model.ReviewAllResponse{Quantity: len(finalClaimedQuests)}, nil
}

func (d *claimedQuestDomain) review(
	ctx context.Context,
	claimedQuests []entity.ClaimedQuest,
	reviewAction entity.ClaimedQuestStatus,
	comment string,
) error {
	if len(claimedQuests) == 0 {
		return errorx.New(errorx.Unavailable, "No claimed quest will be reviewed")
	}

	questSet := map[string]any{}
	claimedQuestSet := map[string]any{}
	for _, cq := range claimedQuests {
		if cq.Status != entity.Pending {
			return errorx.New(errorx.BadRequest, "Claimed quest must be pending")
		}

		claimedQuestSet[cq.ID] = nil
		questSet[cq.QuestID] = nil
	}

	quests, err := d.questRepo.GetByIDs(ctx, common.MapKeys(questSet))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get quest: %v", err)
		return errorx.Unknown
	}

	questInverse := map[string]entity.Quest{}
	for _, q := range quests {
		if q.ProjectID != quests[0].ProjectID {
			return errorx.New(errorx.BadRequest, "You can only review claimed quests of one project")
		}

		questInverse[q.ID] = q
	}

	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	requestUserID := xcontext.RequestUserID(ctx)
	err = d.claimedQuestRepo.UpdateReviewByIDs(ctx, common.MapKeys(claimedQuestSet), &entity.ClaimedQuest{
		Status:     reviewAction,
		Comment:    comment,
		ReviewerID: requestUserID,
		ReviewedAt: time.Now(),
	})

	if err != nil {
		xcontext.Logger(ctx).Errorf("Unable to update status: %v", err)
		return errorx.New(errorx.Internal, "Unable to update status for claim quest")
	}

	for _, claimedQuest := range claimedQuests {
		quest, ok := questInverse[claimedQuest.QuestID]
		if !ok {
			xcontext.Logger(ctx).Errorf("Not found quest %s of claimed quest %s", claimedQuest.QuestID, claimedQuest.ID)
			return errorx.Unknown
		}

		for _, data := range quest.Rewards {
			reward, err := d.questFactory.LoadReward(ctx, quest, data.Type, data.Data)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Invalid reward data: %v", err)
				return errorx.Unknown
			}

			if err := reward.Give(ctx, claimedQuest.UserID, claimedQuest.ID); err != nil {
				return err
			}
		}

		if err := d.increaseTask(ctx, quest.ProjectID.String, claimedQuest.UserID); err != nil {
			xcontext.Logger(ctx).Errorf("Unable to increase number of task: %v", err)
			return errorx.New(errorx.Internal, "Unable to increase number of task")
		}
	}

	xcontext.WithCommitDBTransaction(ctx)
	return nil
}

func (d *claimedQuestDomain) increaseTask(ctx context.Context, projectID, userID string) error {
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
	ctx context.Context, req *model.GiveRewardRequest,
) (*model.GiveRewardResponse, error) {
	if err := d.roleVerifier.Verify(ctx, req.ProjectID, entity.Owner); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denined when give reward: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Only project owner can give reward directly")
	}

	_, err := d.participantRepo.Get(ctx, req.UserID, req.ProjectID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			xcontext.Logger(ctx).Errorf("Cannot get the participant: %v", err)
			return nil, errorx.Unknown
		}

		return nil, errorx.New(errorx.Unavailable, "User must follow the project before")
	}

	rewardType, err := enum.ToEnum[entity.RewardType](req.Type)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid reward type: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid reward type %s", req.Type)
	}

	// Create a fake quest of this project.
	fakeQuest := entity.Quest{ProjectID: sql.NullString{Valid: true, String: req.ProjectID}}
	reward, err := d.questFactory.NewReward(ctx, fakeQuest, rewardType, req.Data)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid reward: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid reward")
	}

	if err := reward.Give(ctx, req.UserID, ""); err != nil {
		xcontext.Logger(ctx).Warnf("Cannot give reward to user: %v", err)
		return nil, errorx.New(errorx.Unavailable, "Cannot give reward to user")
	}

	return &model.GiveRewardResponse{}, nil
}
