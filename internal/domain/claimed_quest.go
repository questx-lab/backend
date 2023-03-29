package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/domain/questutil"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/router"
	"gorm.io/gorm"
)

type ClaimedQuestDomain interface {
	Claim(router.Context, *model.ClaimQuestRequest) (*model.ClaimQuestResponse, error)
	Get(router.Context, *model.GetClaimedQuestRequest) (*model.GetClaimedQuestResponse, error)
	GetList(router.Context, *model.GetListClaimedQuestRequest) (*model.GetListClaimedQuestResponse, error)
}

type claimedQuestDomain struct {
	claimedQuestRepo repository.ClaimedQuestRepository
	questRepo        repository.QuestRepository
}

func NewClaimedQuestDomain(
	claimedQuestRepo repository.ClaimedQuestRepository,
	questRepo repository.QuestRepository,
) *claimedQuestDomain {
	return &claimedQuestDomain{
		claimedQuestRepo: claimedQuestRepo,
		questRepo:        questRepo,
	}
}

func (d *claimedQuestDomain) Claim(
	ctx router.Context, req *model.ClaimQuestRequest,
) (*model.ClaimQuestResponse, error) {
	quest, err := d.questRepo.GetByID(ctx, req.QuestID)
	if err != nil {
		ctx.Logger().Errorf("Cannot get quest: %v", err)
		return nil, errorx.Unknown
	}

	if quest.Status != entity.Published {
		return nil, errorx.New(errorx.Unavailable, "Only allowed to claim published quest")
	}

	claimable, err := d.isClaimable(ctx, *quest)
	if err != nil {
		ctx.Logger().Errorf("Cannot determine claimable: %v", err)
		return nil, errorx.Unknown
	}

	if !claimable {
		return nil, errorx.New(errorx.Unavailable, "This quest cannot be claimed now")
	}

	validatorFactory := questutil.NewValidatorFactory()
	validator, err := validatorFactory.New(ctx, quest.Type, quest.ValidationData)
	if err != nil {
		ctx.Logger().Debugf("Invalid validation data: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid validation data")
	}

	var status entity.ClaimedQuestStatus

	success, err := validator.Validate(ctx, req.Input)
	switch {
	case err != nil:
		var errx errorx.Error
		if !errors.As(err, &errx) {
			ctx.Logger().Errorf("Got an invalid error when validate claimed quest: %v", err)
			return nil, errorx.Unknown
		}

		if errx.Code != errorx.NeedManualReview {
			return nil, errx
		}

		status = entity.Pending

	case success:
		status = entity.AutoAccepted

	case !success:
		status = entity.Rejected
	}

	claimedQuest := &entity.ClaimedQuest{
		Base:    entity.Base{ID: uuid.NewString()},
		QuestID: req.QuestID,
		UserID:  ctx.GetUserID(),
		Status:  status,
		Input:   req.Input,
	}

	if status != entity.Pending {
		claimedQuest.ReviewerAt = time.Now()
	}

	// Only save the rejected claimed quest if it's a manual reviewed quest.
	if status != entity.Rejected {
		err = d.claimedQuestRepo.Create(ctx, claimedQuest)
		if err != nil {
			ctx.Logger().Errorf("Cannot claim quest: %v", err)
			return nil, errorx.Unknown
		}
	}

	if status == entity.AutoAccepted {
		awardFactory := questutil.NewAwardFactory()
		for _, data := range quest.Awards {
			award, err := awardFactory.New(ctx, data)
			if err != nil {
				ctx.Logger().Errorf("Invalid award data: %v", err)
				return nil, errorx.Unknown
			}

			if err := award.Give(ctx); err != nil {
				return nil, err
			}
		}
	}

	return &model.ClaimQuestResponse{ID: claimedQuest.ID, Status: string(status)}, nil
}

func (d *claimedQuestDomain) Get(
	ctx router.Context, req *model.GetClaimedQuestRequest,
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
	ctx router.Context, req *model.GetListClaimedQuestRequest,
) (*model.GetListClaimedQuestResponse, error) {
	if req.ProjectID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty project id")
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

	result, err := d.claimedQuestRepo.GetList(ctx, req.ProjectID, req.Offset, req.Limit)
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

func (d *claimedQuestDomain) isClaimable(ctx router.Context, quest entity.Quest) (bool, error) {
	// Check conditions.
	conditionFactory := questutil.NewConditionFactory(d.claimedQuestRepo, d.questRepo)
	for _, c := range quest.Conditions {
		condition, err := conditionFactory.New(ctx, c)
		if err != nil {
			return false, err
		}

		b, err := condition.Check(ctx)
		if err != nil {
			return false, err
		}

		if !b {
			return false, nil
		}
	}

	// Check recurrence.
	lastClaimedQuest, err := d.claimedQuestRepo.GetLastPendingOrAccepted(ctx, ctx.GetUserID(), quest.ID)
	if err != nil {
		// The user has not yet claimed this quest.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil
		}
		return false, err
	}

	// The user claimed the quest before.
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
