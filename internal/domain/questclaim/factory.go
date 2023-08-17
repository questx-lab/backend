package questclaim

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/api/discord"
	"github.com/questx-lab/backend/pkg/api/telegram"
	"github.com/questx-lab/backend/pkg/api/twitter"
	"github.com/questx-lab/backend/pkg/dateutil"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

const day = 24 * time.Hour
const week = 7 * day

var referralReward Reward
var referralRewardMutex sync.Mutex

type Factory struct {
	claimedQuestRepo repository.ClaimedQuestRepository
	questRepo        repository.QuestRepository
	communityRepo    repository.CommunityRepository
	followerRepo     repository.FollowerRepository
	oauth2Repo       repository.OAuth2Repository
	userRepo         repository.UserRepository
	payRewardRepo    repository.PayRewardRepository
	blockchainRepo   repository.BlockChainRepository
	lotteryRepo      repository.LotteryRepository

	twitterEndpoint  twitter.IEndpoint
	discordEndpoint  discord.IEndpoint
	telegramEndpoint telegram.IEndpoint

	blockchainCaller client.BlockchainCaller
}

func NewFactory(
	claimedQuestRepo repository.ClaimedQuestRepository,
	questRepo repository.QuestRepository,
	communityRepo repository.CommunityRepository,
	followerRepo repository.FollowerRepository,
	oauth2Repo repository.OAuth2Repository,
	userRepo repository.UserRepository,
	payRewardRepo repository.PayRewardRepository,
	blockchainRepo repository.BlockChainRepository,
	lotteryRepo repository.LotteryRepository,
	twitterEndpoint twitter.IEndpoint,
	discordEndpoint discord.IEndpoint,
	telegramEndpoint telegram.IEndpoint,
	blockchainCaller client.BlockchainCaller,
) Factory {
	return Factory{
		claimedQuestRepo: claimedQuestRepo,
		questRepo:        questRepo,
		communityRepo:    communityRepo,
		followerRepo:     followerRepo,
		oauth2Repo:       oauth2Repo,
		userRepo:         userRepo,
		payRewardRepo:    payRewardRepo,
		blockchainRepo:   blockchainRepo,
		lotteryRepo:      lotteryRepo,
		twitterEndpoint:  twitterEndpoint,
		discordEndpoint:  discordEndpoint,
		telegramEndpoint: telegramEndpoint,
		blockchainCaller: blockchainCaller,
	}
}

// NewProcessor creates a new processor and validate the data.
func (f Factory) NewProcessor(ctx context.Context, quest entity.Quest, data map[string]any) (Processor, error) {
	return f.newProcessor(ctx, quest, data, true, true)
}

// LoadProcessor creates a new processor but not validate the data.
func (f Factory) LoadProcessor(ctx context.Context, includeSecret bool, quest entity.Quest, data map[string]any) (Processor, error) {
	return f.newProcessor(ctx, quest, data, false, includeSecret)
}

func (f Factory) newProcessor(
	ctx context.Context,
	quest entity.Quest,
	data map[string]any,
	needParse, includeSecret bool,
) (Processor, error) {
	var processor Processor
	var err error

	switch quest.Type {
	case entity.QuestVisitLink:
		processor, err = newVisitLinkProcessor(ctx, data, needParse)

	case entity.QuestText:
		processor, err = newTextProcessor(ctx, data, needParse, includeSecret)

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
	ctx context.Context,
	quest entity.Quest,
	conditionType entity.ConditionType,
	data map[string]any,
) (Condition, error) {
	return f.newCondition(ctx, quest, conditionType, data, true)
}

// LoadCondition creates a new condition but not validate the data.
func (f Factory) LoadCondition(
	ctx context.Context,
	quest entity.Quest,
	conditionType entity.ConditionType,
	data map[string]any,
) (Condition, error) {
	return f.newCondition(ctx, quest, conditionType, data, false)
}

func (f Factory) newCondition(
	ctx context.Context,
	quest entity.Quest,
	conditionType entity.ConditionType,
	data map[string]any,
	needParse bool,
) (Condition, error) {
	var condition Condition
	var err error
	switch conditionType {
	case entity.QuestCondition:
		condition, err = NewQuestCondition(ctx, f, data, needParse)

	case entity.DateCondition:
		condition, err = newDateCondition(ctx, data, needParse)

	case entity.DiscordCondition:
		condition, err = newDiscordCondition(ctx, f, quest, data, needParse)

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
	ctx context.Context,
	communityID string,
	rewardType entity.RewardType,
	data map[string]any,
) (Reward, error) {
	return f.newReward(ctx, communityID, rewardType, data, true)
}

// LoadReward creates a new reward but not validate the data.
func (f Factory) LoadReward(
	ctx context.Context,
	communityID string,
	rewardType entity.RewardType,
	data map[string]any,
) (Reward, error) {
	return f.newReward(ctx, communityID, rewardType, data, false)
}

func (f Factory) newReward(
	ctx context.Context,
	communityID string,
	rewardType entity.RewardType,
	data map[string]any,
	needParse bool,
) (Reward, error) {
	var reward Reward
	var err error
	switch rewardType {
	case entity.DiscordRoleReward:
		reward, err = newDiscordRoleReward(ctx, communityID, f, data, needParse)

	case entity.CoinReward:
		reward, err = newCoinReward(ctx, f, data, needParse)

	case entity.NFTReward:
		reward, err = newNonFungibleTokenReward(ctx, f, data, needParse)

	default:
		return nil, fmt.Errorf("invalid reward type %s", rewardType)
	}

	if err != nil {
		return nil, err
	}

	return reward, nil
}

func (f Factory) getRequestServiceUserID(ctx context.Context, service string) string {
	serviceUser, err := f.oauth2Repo.GetByUserID(ctx, service, xcontext.RequestUserID(ctx))
	if err != nil {
		return ""
	}

	if service == xcontext.Configs(ctx).Auth.Twitter.Name {
		return serviceUser.ServiceUsername
	}

	serviceName, id, found := strings.Cut(serviceUser.ServiceUserID, "_")
	if !found || serviceName != service {
		return ""
	}

	if service == xcontext.Configs(ctx).Auth.Twitter.Name {
		return serviceUser.ServiceUsername
	}

	return id
}

type UnclaimableReasonType int

const (
	UnclaimableByUnknown UnclaimableReasonType = iota
	UnclaimableByRetryAfter
	UnclaimableByCondition
	UnclaimableByRecurrence
)

type UnclaimableReason struct {
	Type     UnclaimableReasonType
	Message  string
	Metadata map[string]any
}

func (f Factory) IsClaimable(ctx context.Context, quest entity.Quest) (*UnclaimableReason, error) {
	// Check time for reclaiming.
	lastRejectedClaimedQuest, err := f.claimedQuestRepo.GetLast(
		ctx,
		repository.GetLastClaimedQuestFilter{
			UserID:  xcontext.RequestUserID(ctx),
			QuestID: quest.ID,
			Status:  []entity.ClaimedQuestStatus{entity.AutoRejected, entity.Rejected},
		},
	)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	processor, err := f.LoadProcessor(ctx, true, quest, quest.ValidationData)
	if err != nil {
		return nil, err
	}

	retryAfter := processor.RetryAfter()
	if retryAfter != 0 && lastRejectedClaimedQuest != nil {
		lastRejectedAt := lastRejectedClaimedQuest.CreatedAt
		if elapsed := time.Since(lastRejectedAt); elapsed <= retryAfter {
			waitFor := retryAfter - elapsed
			waitFor = waitFor / time.Second * time.Second // Remove nanosecond from duration.
			return &UnclaimableReason{
				Type:    UnclaimableByRetryAfter,
				Message: fmt.Sprintf("Please wait for %s before continuing to claim this quest", waitFor),
			}, nil
		}
	}

	// Check conditions.
	finalCondition := true
	if quest.ConditionOp == entity.Or && len(quest.Conditions) > 0 {
		finalCondition = false
	}

	var firstFailedCondition Condition
	for _, c := range quest.Conditions {
		condition, err := f.LoadCondition(ctx, quest, c.Type, c.Data)
		if err != nil {
			return &UnclaimableReason{Type: UnclaimableByUnknown}, err
		}

		ok, err := condition.Check(ctx)
		if err != nil {
			return &UnclaimableReason{Type: UnclaimableByUnknown}, err
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
		return &UnclaimableReason{
			Type:    UnclaimableByCondition,
			Message: firstFailedCondition.Statement(),
		}, nil
	}

	// Check recurrence.
	lastClaimedQuest, err := f.claimedQuestRepo.GetLast(
		ctx,
		repository.GetLastClaimedQuestFilter{
			UserID:  xcontext.RequestUserID(ctx),
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
			return nil, nil
		}

		return &UnclaimableReason{Type: UnclaimableByUnknown}, err
	}

	// If the user claimed the quest before, this quest cannot be claimed again until the next
	// recurrence.
	lastClaimedAt := lastClaimedQuest.CreatedAt
	switch quest.Recurrence {
	case entity.Once:
		return &UnclaimableReason{
			Type:    UnclaimableByRecurrence,
			Message: "recurrence",
		}, nil

	case entity.Daily:
		if lastClaimedAt.Day() != time.Now().Day() || time.Since(lastClaimedAt) > day {
			return nil, nil
		}

		return &UnclaimableReason{
			Type:     UnclaimableByRecurrence,
			Message:  "recurrence",
			Metadata: map[string]any{"next_claim": dateutil.NextDay(time.Now())},
		}, nil

	case entity.Weekly:
		_, lastWeek := lastClaimedAt.ISOWeek()
		_, currentWeek := time.Now().ISOWeek()
		if lastWeek != currentWeek || time.Since(lastClaimedAt) > week {
			return nil, nil
		}

		return &UnclaimableReason{
			Type:     UnclaimableByRecurrence,
			Message:  "recurrence",
			Metadata: map[string]any{"next_claim": dateutil.NextWeek(time.Now())},
		}, nil

	case entity.Monthly:
		if lastClaimedAt.Month() != time.Now().Month() || lastClaimedAt.Year() != time.Now().Year() {
			return nil, nil
		}

		return &UnclaimableReason{
			Type:     UnclaimableByRecurrence,
			Message:  "recurrence",
			Metadata: map[string]any{"next_claim": dateutil.NextMonth(time.Now())},
		}, nil

	default:
		return &UnclaimableReason{Type: UnclaimableByUnknown}, fmt.Errorf("invalid recurrence %s", quest.Recurrence)
	}
}

func (f Factory) LoadReferralReward(ctx context.Context) (Reward, error) {
	if referralReward == nil {
		referralRewardMutex.Lock()
		defer referralRewardMutex.Unlock()

		if referralReward == nil {
			reward, err := newCoinReward(ctx, f, map[string]any{
				"chain":         xcontext.Configs(ctx).Quest.InviteCommunityRewardChain,
				"token_address": xcontext.Configs(ctx).Quest.InviteCommunityRewardTokenAddress,
				"amount":        xcontext.Configs(ctx).Quest.InviteCommunityRewardAmount,
			}, true)
			if err != nil {
				return nil, err
			}

			referralReward = reward
		}
	}

	return referralReward, nil
}
