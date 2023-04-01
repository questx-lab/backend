package questclaim

import (
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/xcontext"
)

// Processor Factory
func NewProcessor(ctx xcontext.Context, t entity.QuestType, data string) (Processor, error) {
	var processor Processor
	var err error
	switch t {
	case entity.VisitLink:
		processor, err = newVisitLinkProcessor(ctx, data)

	case entity.Text:
		processor, err = newTextProcessor(ctx, data)

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
	participantRepo repository.ParticipantRepository,
	data entity.Award,
) (Award, error) {
	var award Award
	var err error
	switch data.Type {
	case entity.PointAward:
		award, err = newPointAward(ctx, participantRepo, data)

	case entity.DiscordRole:
		award, err = newDiscordRoleAward(ctx, data)

	default:
		return nil, fmt.Errorf("invalid award type %s", data.Type)
	}

	if err != nil {
		return nil, err
	}

	return award, nil
}
