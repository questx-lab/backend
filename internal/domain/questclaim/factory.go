package questclaim

import (
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/router"
)

// Validator Factory
func NewValidator(ctx router.Context, t entity.QuestType, data string) (Validator, error) {
	var validator Validator
	var err error
	switch t {
	case entity.VisitLink:
		validator, err = newVisitLinkValidator(ctx, data)

	case entity.Text:
		validator, err = newTextValidator(ctx, data)

	default:
		return nil, fmt.Errorf("invalid quest type %s", t)
	}

	if err != nil {
		return nil, err
	}

	return validator, nil
}

// Condition Factory
func NewCondition(
	ctx router.Context,
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
func NewAward(ctx router.Context, data entity.Award) (Award, error) {
	var award Award
	var err error
	switch data.Type {
	case entity.PointAward:
		award, err = newPointAward(ctx, data)

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
