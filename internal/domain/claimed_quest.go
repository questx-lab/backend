package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/questclaim"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

type ClaimedQuestDomain interface {
	Claim(xcontext.Context, *model.ClaimQuestRequest) (*model.ClaimQuestResponse, error)
	Get(xcontext.Context, *model.GetClaimedQuestRequest) (*model.GetClaimedQuestResponse, error)
	GetList(xcontext.Context, *model.GetListClaimedQuestRequest) (*model.GetListClaimedQuestResponse, error)
	GetPendingList(xcontext.Context, *model.GetPendingListClaimedQuestRequest) (*model.GetPendingListClaimedQuestResponse, error)
	ReviewClaimedQuest(xcontext.Context, *model.ReviewClaimedQuestRequest) (*model.ReviewClaimedQuestResponse, error)
}

type claimedQuestDomain struct {
	claimedQuestRepo repository.ClaimedQuestRepository
	questRepo        repository.QuestRepository
	participantRepo  repository.ParticipantRepository
	roleVerifier     *common.ProjectRoleVerifier
}

func NewClaimedQuestDomain(
	claimedQuestRepo repository.ClaimedQuestRepository,
	questRepo repository.QuestRepository,
	collaboratorRepo repository.CollaboratorRepository,
	participantRepo repository.ParticipantRepository,
) *claimedQuestDomain {
	return &claimedQuestDomain{
		claimedQuestRepo: claimedQuestRepo,
		questRepo:        questRepo,
		participantRepo:  participantRepo,
		roleVerifier:     common.NewProjectRoleVerifier(collaboratorRepo),
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

	// Check if user joins in project.
	_, err = d.participantRepo.Get(ctx, xcontext.GetRequestUserID(ctx), quest.ProjectID)
	if err != nil {
		return nil, errorx.New(errorx.PermissionDenied, "You have not joined the project yet")
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
	processor, err := questclaim.NewProcessor(ctx, quest.Type, quest.ValidationData)
	if err != nil {
		ctx.Logger().Debugf("Invalid validation data: %v", err)
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
		UserID:  xcontext.GetRequestUserID(ctx),
		Status:  status,
		Input:   req.Input,
	}

	if status != entity.Pending {
		claimedQuest.ReviewerAt = time.Now()
	}

	// GiveAward can write something to database.
	ctx.BeginTx()
	defer ctx.RollbackTx()

	err = d.claimedQuestRepo.Create(ctx, claimedQuest)
	if err != nil {
		ctx.Logger().Errorf("Cannot claim quest: %v", err)
		return nil, errorx.Unknown
	}

	// Give award to user if the claimed quest is accepted.
	if status == entity.AutoAccepted {
		for _, data := range quest.Awards {
			award, err := questclaim.NewAward(ctx, d.participantRepo, data)
			if err != nil {
				ctx.Logger().Errorf("Invalid award data: %v", err)
				return nil, errorx.Unknown
			}

			if err := award.Give(ctx, quest.ProjectID); err != nil {
				return nil, err
			}
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

	if err := d.roleVerifier.Verify(ctx, req.ProjectID, entity.AdminGroup...); err != nil {
		ctx.Logger().Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	if req.Limit == 0 {
		req.Limit = 1
	}

	if req.Limit < 0 {
		return nil, errorx.New(errorx.BadRequest, "Limit must be positive")
	}

	if req.Limit > 50 {
		return nil, errorx.New(errorx.BadRequest, "Exceed the maximum of limit")
	}

	result, err := d.claimedQuestRepo.GetList(ctx, &repository.ClaimedQuestFilter{
		ProjectID: req.ProjectID,
	}, req.Offset, req.Limit)
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
		condition, err := questclaim.NewCondition(ctx, d.claimedQuestRepo, d.questRepo, c)
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
	userID := xcontext.GetRequestUserID(ctx)
	lastClaimedQuest, err := d.claimedQuestRepo.GetLastPendingOrAccepted(ctx, userID, quest.ID)
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

	if !slices.Contains([]entity.ClaimedQuestStatus{entity.Accepted, entity.Rejected}, entity.ClaimedQuestStatus(req.Action)) {
		return nil, errorx.New(errorx.BadRequest, "Status must be accept or reject")
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

	userID := xcontext.GetRequestUserID(ctx)
	if err := d.claimedQuestRepo.UpdateReviewByID(ctx, req.ID, &entity.ClaimedQuest{
		Status:     entity.ClaimedQuestStatus(req.Action),
		ReviewerID: userID,
		ReviewerAt: time.Now(),
	}); err != nil {
		ctx.Logger().Errorf("Unable to update status: %v", err)
		return nil, errorx.New(errorx.Internal, "Unable to approve this claim quest")
	}

	return &model.ReviewClaimedQuestResponse{}, nil
}

func (d *claimedQuestDomain) GetPendingList(ctx xcontext.Context, req *model.GetPendingListClaimedQuestRequest) (*model.GetPendingListClaimedQuestResponse, error) {
	if req.ProjectID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty project id")
	}

	if err := d.roleVerifier.Verify(ctx, req.ProjectID, entity.ReviewGroup...); err != nil {
		ctx.Logger().Errorf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	if req.Limit == 0 {
		req.Limit = 1
	}

	if req.Limit < 0 {
		return nil, errorx.New(errorx.BadRequest, "Limit must be positive")
	}

	if req.Limit > 50 {
		return nil, errorx.New(errorx.BadRequest, "Exceed the maximum of limit")
	}

	result, err := d.claimedQuestRepo.GetList(ctx, &repository.ClaimedQuestFilter{
		ProjectID: req.ProjectID,
		Status:    entity.Pending,
	}, req.Offset, req.Limit)
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

	return &model.GetPendingListClaimedQuestResponse{ClaimedQuests: claimedQuests}, nil
}
