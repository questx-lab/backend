package questclaim

import (
	"encoding/json"
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/api/discord"
	"github.com/questx-lab/backend/pkg/api/twitter"
	"github.com/questx-lab/backend/pkg/xcontext"
)

// Processor Factory
func NewProcessor(
	ctx xcontext.Context,
	quest entity.Quest,
	projectRepo repository.ProjectRepository,
	twitterEndpoint twitter.IEndpoint,
	discordEndpoint discord.IEndpoint,
	t entity.QuestType,
	data any,
) (Processor, error) {
	mapdata := map[string]any{}
	switch t := data.(type) {
	case []byte:
		err := json.Unmarshal(t, &mapdata)
		if err != nil {
			return nil, err
		}
	case map[string]any:
		mapdata = t
	default:
		return nil, fmt.Errorf("invalid data type %T", data)
	}

	var processor Processor
	var err error
	switch t {
	case entity.QuestVisitLink:
		processor, err = newVisitLinkProcessor(ctx, mapdata)

	case entity.QuestText:
		processor, err = newTextProcessor(ctx, mapdata)

	case entity.QuestTwitterFollow:
		processor, err = newTwitterFollowProcessor(ctx, twitterEndpoint, mapdata)

	case entity.QuestTwitterReaction:
		processor, err = newTwitterReactionProcessor(ctx, twitterEndpoint, mapdata)

	case entity.QuestTwitterTweet:
		processor, err = newTwitterTweetProcessor(ctx, twitterEndpoint, mapdata)

	case entity.QuestTwitterJoinSpace:
		processor, err = newTwitterJoinSpaceProcessor(ctx, twitterEndpoint, mapdata)

	case entity.QuestJoinDiscord:
		processor, err = newJoinDiscordProcessor(ctx, projectRepo, quest, discordEndpoint, mapdata)

	default:
		return nil, fmt.Errorf("invalid quest type %s", t)
	}

	if err != nil {
		return nil, err
	}

	return processor, nil
}

// Condition Factory
func NewCondition(
	ctx xcontext.Context,
	claimedQuestRepo repository.ClaimedQuestRepository,
	questRepo repository.QuestRepository,
	data entity.Condition,
) (Condition, error) {
	var condition Condition
	var err error
	switch data.Type {
	case entity.QuestCondition:
		condition, err = newQuestCondition(ctx, data, claimedQuestRepo, questRepo)

	case entity.DateCondition:
		condition, err = newDateCondition(ctx, data)

	default:
		return nil, fmt.Errorf("invalid condition type %s", data.Type)
	}

	if err != nil {
		return nil, err
	}

	return condition, nil
}

// Award Factory
func NewAward(
	ctx xcontext.Context,
	quest entity.Quest,
	projectRepo repository.ProjectRepository,
	participantRepo repository.ParticipantRepository,
	discordEndpoint discord.IEndpoint,
	data entity.Award,
) (Award, error) {
	var award Award
	var err error
	switch data.Type {
	case entity.PointAward:
		award, err = newPointAward(ctx, quest, participantRepo, data)

	case entity.DiscordRole:
		award, err = newDiscordRoleAward(ctx, quest, projectRepo, discordEndpoint, data)

	default:
		return nil, fmt.Errorf("invalid award type %s", data.Type)
	}

	if err != nil {
		return nil, err
	}

	return award, nil
}
