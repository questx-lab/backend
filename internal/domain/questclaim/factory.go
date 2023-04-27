package questclaim

import (
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/api/discord"
	"github.com/questx-lab/backend/pkg/api/twitter"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type Factory struct {
	claimedQuestRepo repository.ClaimedQuestRepository
	questRepo        repository.QuestRepository
	projectRepo      repository.ProjectRepository
	participantRepo  repository.ParticipantRepository

	twitterEndpoint twitter.IEndpoint
	discordEndpoint discord.IEndpoint
}

func NewFactory(
	claimedQuestRepo repository.ClaimedQuestRepository,
	questRepo repository.QuestRepository,
	projectRepo repository.ProjectRepository,
	participantRepo repository.ParticipantRepository,
	twitterEndpoint twitter.IEndpoint,
	discordEndpoint discord.IEndpoint,
) Factory {
	return Factory{
		claimedQuestRepo: claimedQuestRepo,
		questRepo:        questRepo,
		projectRepo:      projectRepo,
		participantRepo:  participantRepo,
		twitterEndpoint:  twitterEndpoint,
		discordEndpoint:  discordEndpoint,
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

	case entity.QuestJoinTelegram:
		processor, err = newJoinTelegramProcessor(ctx, data)

	case entity.QuestInvite:
		processor, err = newInviteProcessor(ctx, data, needParse)

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
