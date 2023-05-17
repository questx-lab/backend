package questclaim

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/api/discord"
	"github.com/questx-lab/backend/pkg/api/telegram"
	"github.com/questx-lab/backend/pkg/api/twitter"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

const errReason = "Something is wrong"
const passReason = ""
const day = 24 * time.Hour
const week = 7 * day

type Factory struct {
	claimedQuestRepo  repository.ClaimedQuestRepository
	questRepo         repository.QuestRepository
	projectRepo       repository.ProjectRepository
	participantRepo   repository.ParticipantRepository
	oauth2Repo        repository.OAuth2Repository
	userAggregateRepo repository.UserAggregateRepository

	twitterEndpoint  twitter.IEndpoint
	discordEndpoint  discord.IEndpoint
	telegramEndpoint telegram.IEndpoint

	projectRoleVerifier *common.ProjectRoleVerifier
}

func NewFactory(
	claimedQuestRepo repository.ClaimedQuestRepository,
	questRepo repository.QuestRepository,
	projectRepo repository.ProjectRepository,
	participantRepo repository.ParticipantRepository,
	oauth2Repo repository.OAuth2Repository,
	userAggregateRepo repository.UserAggregateRepository,
	projectRoleVerifier *common.ProjectRoleVerifier,
	twitterEndpoint twitter.IEndpoint,
	discordEndpoint discord.IEndpoint,
	telegramEndpoint telegram.IEndpoint,
) Factory {
	return Factory{
		claimedQuestRepo:    claimedQuestRepo,
		questRepo:           questRepo,
		projectRepo:         projectRepo,
		participantRepo:     participantRepo,
		oauth2Repo:          oauth2Repo,
		userAggregateRepo:   userAggregateRepo,
		twitterEndpoint:     twitterEndpoint,
		discordEndpoint:     discordEndpoint,
		telegramEndpoint:    telegramEndpoint,
		projectRoleVerifier: projectRoleVerifier,
	}
}

// NewProcessor creates a new processor and validate the data.
func (f Factory) NewProcessor(ctx xcontext.Context, quest entity.Quest, data map[string]any) (Processor, error) {
	return f.newProcessor(ctx, quest, data, true)
}

// LoadProcessor creates a new processor but not validate the data.
func (f Factory) LoadProcessor(ctx xcontext.Context, quest entity.Quest, data map[string]any) (Processor, error) {
	return f.newProcessor(ctx, quest, data, false)
}

func (f Factory) newProcessor(
	ctx xcontext.Context,
	quest entity.Quest,
	data map[string]any,
	needParse bool,
) (Processor, error) {
	var processor Processor
	var err error

	switch quest.Type {
	case entity.QuestVisitLink:
		processor, err = newVisitLinkProcessor(ctx, data, needParse)

	case entity.QuestText:
		processor, err = newTextProcessor(ctx, data, needParse)

	case entity.QuestQuiz:
		processor, err = newQuizProcessor(ctx, data, needParse)
	case entity.QuestEmpty:
		processor, err = newEmptyProcessor(ctx, data)

	case entity.QuestURL:
		processor, err = newURLProcessor(ctx, data)

	case entity.QuestImage:
		processor, err = newImageProcessor(ctx, data)

	case entity.QuestTwitterFollow:
		processor, err = newTwitterFollowProcessor(ctx, f, data, needParse)

	case entity.QuestTwitterReaction:
		processor, err = newTwitterReactionProcessor(ctx, f, data, needParse)

	case entity.QuestTwitterTweet:
		processor, err = newTwitterTweetProcessor(ctx, f, data)

	case entity.QuestTwitterJoinSpace:
		processor, err = newTwitterJoinSpaceProcessor(ctx, f, data)

	case entity.QuestJoinDiscord:
		processor, err = newJoinDiscordProcessor(ctx, f, quest, data, needParse)

	case entity.QuestInviteDiscord:
		processor, err = newInviteDiscordProcessor(ctx, f, quest, data, needParse)

	case entity.QuestJoinTelegram:
		processor, err = newJoinTelegramProcessor(ctx, f, quest, data, needParse)

	case entity.QuestInvite:
		processor, err = newInviteProcessor(ctx, f, quest, data, needParse)

	default:
		return nil, fmt.Errorf("invalid quest type %s", quest.Type)
	}

	if err != nil {
		return nil, err
	}

	return processor, nil
}

// NewCondition creates a new condition and validate the data.
func (f Factory) NewCondition(
	ctx xcontext.Context,
	conditionType entity.ConditionType,
	data map[string]any,
) (Condition, error) {
	return f.newCondition(ctx, conditionType, data, true)
}

// LoadCondition creates a new condition but not validate the data.
func (f Factory) LoadCondition(
	ctx xcontext.Context,
	conditionType entity.ConditionType,
	data map[string]any,
) (Condition, error) {
	return f.newCondition(ctx, conditionType, data, false)
}

func (f Factory) newCondition(
	ctx xcontext.Context,
	conditionType entity.ConditionType,
	data map[string]any,
	needParse bool,
) (Condition, error) {
	var condition Condition
	var err error
	switch conditionType {
	case entity.QuestCondition:
		condition, err = newQuestCondition(ctx, f, data, needParse)

	case entity.DateCondition:
		condition, err = newDateCondition(ctx, data, needParse)

	default:
		return nil, fmt.Errorf("invalid condition type %s", conditionType)
	}

	if err != nil {
		return nil, err
	}

	return condition, nil
}

// NewReward creates a new reward and validate the data.
func (f Factory) NewReward(
	ctx xcontext.Context,
	quest entity.Quest,
	rewardType entity.RewardType,
	data map[string]any,
) (Reward, error) {
	return f.newReward(ctx, quest, rewardType, data, true)
}

// LoadReward creates a new reward but not validate the data.
func (f Factory) LoadReward(
	ctx xcontext.Context,
	quest entity.Quest,
	rewardType entity.RewardType,
	data map[string]any,
) (Reward, error) {
	return f.newReward(ctx, quest, rewardType, data, false)
}

func (f Factory) newReward(
	ctx xcontext.Context,
	quest entity.Quest,
	rewardType entity.RewardType,
	data map[string]any,
	needParse bool,
) (Reward, error) {
	var reward Reward
	var err error
	switch rewardType {
	case entity.PointReward:
		reward, err = newPointReward(ctx, quest, f, data)

	case entity.DiscordRole:
		reward, err = newDiscordRoleReward(ctx, quest, f, data, needParse)

	default:
		return nil, fmt.Errorf("invalid reward type %s", rewardType)
	}

	if err != nil {
		return nil, err
	}

	return reward, nil
}

func (f Factory) getRequestServiceUserID(ctx xcontext.Context, service string) string {
	serviceUser, err := f.oauth2Repo.GetByUserID(ctx, service, xcontext.GetRequestUserID(ctx))
	if err != nil {
		return ""
	}

	serviceName, id, found := strings.Cut(serviceUser.ServiceUserID, "_")
	if !found || serviceName != service {
		return ""
	}

	return id
}

func (f Factory) IsClaimable(ctx xcontext.Context, quest entity.Quest) (reason string, err error) {
	// Check time for reclaiming.
	lastRejectedClaimedQuest, err := f.claimedQuestRepo.GetLast(
		ctx,
		repository.GetLastClaimedQuestFilter{
			UserID:  xcontext.GetRequestUserID(ctx),
			QuestID: quest.ID,
			Status:  []entity.ClaimedQuestStatus{entity.AutoRejected, entity.Rejected},
		},
	)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}

	processor, err := f.LoadProcessor(ctx, quest, quest.ValidationData)
	if err != nil {
		return "", err
	}

	retryAfter := processor.RetryAfter()
	if retryAfter != 0 && err == nil {
		lastRejectedAt := lastRejectedClaimedQuest.CreatedAt
		if elapsed := time.Since(lastRejectedAt); elapsed <= retryAfter {
			waitFor := retryAfter - elapsed
			return fmt.Sprintf("Please wait for %s before continuing to claim this quest", waitFor), nil
		}
	}

	// Check conditions.
	finalCondition := true
	if quest.ConditionOp == entity.Or && len(quest.Conditions) > 0 {
		finalCondition = false
	}

	var firstFailedCondition Condition
	for _, c := range quest.Conditions {
		condition, err := f.LoadCondition(ctx, c.Type, c.Data)
		if err != nil {
			return errReason, err
		}

		ok, err := condition.Check(ctx)
		if err != nil {
			return errReason, err
		}

		if firstFailedCondition == nil && !ok {
			firstFailedCondition = condition
		}

		if quest.ConditionOp == entity.And {
			finalCondition = finalCondition && ok
		} else {
			finalCondition = finalCondition || ok
		}
	}

	if !finalCondition {
		return firstFailedCondition.Statement(), nil
	}

	// Check recurrence.
	lastClaimedQuest, err := f.claimedQuestRepo.GetLast(
		ctx,
		repository.GetLastClaimedQuestFilter{
			UserID:  xcontext.GetRequestUserID(ctx),
			QuestID: quest.ID,
			Status: []entity.ClaimedQuestStatus{
				entity.Pending,
				entity.Accepted,
				entity.AutoAccepted,
			},
		},
	)
	if err != nil {
		// The user has not claimed this quest yet.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return passReason, nil
		}

		return errReason, err
	}

	// If the user claimed the quest before, this quest cannot be claimed again until the next
	// recurrence.
	lastClaimedAt := lastClaimedQuest.CreatedAt
	switch quest.Recurrence {
	case entity.Once:
		return "This quest can only claim once", nil

	case entity.Daily:
		if lastClaimedAt.Day() != time.Now().Day() || time.Since(lastClaimedAt) > day {
			return passReason, nil
		}

		return "Please wait until the next day to claim this quest", nil

	case entity.Weekly:
		_, lastWeek := lastClaimedAt.ISOWeek()
		_, currentWeek := time.Now().ISOWeek()
		if lastWeek != currentWeek || time.Since(lastClaimedAt) > week {
			return passReason, nil
		}

		return "Please wait until the next week to claim this quest", nil

	case entity.Monthly:

		if lastClaimedAt.Month() != time.Now().Month() || lastClaimedAt.Year() != time.Now().Year() {
			return passReason, nil
		}

		return "Please wait until the next month to claim this quest", nil

	default:
		return errReason, fmt.Errorf("invalid recurrence %s", quest.Recurrence)
	}
}
