package gameengine

import (
	"errors"
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/enum"
)

const (
	MoveActionType = "move"
	JoinActionType = "join"
	ExitActionType = "exit"
	InitActionType = "init"
)

type Action interface {
	// OnlyOwner indicates that the response of this action will only send to
	// the owner of this action or broadcast to all clients.
	OnlyOwner() bool

	Apply(*GameState) error
}

func formatAction(id int, a Action) (model.GameActionResponse, error) {
	switch t := a.(type) {
	case *MoveAction:
		return model.GameActionResponse{
			ID:        id,
			UserID:    t.UserID,
			OnlyOwner: t.OnlyOwner(),
			Type:      MoveActionType,
			Value:     map[string]any{"direction": t.Direction},
		}, nil

	case *JoinAction:
		return model.GameActionResponse{
			ID:        id,
			Type:      JoinActionType,
			OnlyOwner: t.OnlyOwner(),
			UserID:    t.UserID,
			Value: map[string]any{
				"position":  t.position,
				"direction": t.direction,
			},
		}, nil

	case *ExitAction:
		return model.GameActionResponse{
			ID:        id,
			UserID:    t.UserID,
			OnlyOwner: t.OnlyOwner(),
			Type:      ExitActionType,
			Value:     nil,
		}, nil

	case *InitAction:
		return model.GameActionResponse{
			ID:        id,
			UserID:    t.UserID,
			OnlyOwner: t.OnlyOwner(),
			Type:      InitActionType,
			Value:     map[string]any{"users": t.initialUsers},
		}, nil

	default:
		return model.GameActionResponse{}, fmt.Errorf("not set up action %T", a)
	}
}

func parseAction(req model.GameActionServerRequest) (Action, error) {
	switch req.Type {
	case MoveActionType:
		direction, ok := req.Value["direction"].(string)
		if !ok {
			return nil, errors.New("invalid or not found direction")
		}

		directionEnum, err := enum.ToEnum[entity.DirectionType](direction)
		if err != nil {
			return nil, err
		}

		return &MoveAction{
			UserID:    req.UserID,
			Direction: directionEnum,
		}, nil

	case JoinActionType:
		return &JoinAction{UserID: req.UserID}, nil

	case ExitActionType:
		return &ExitAction{UserID: req.UserID}, nil

	case InitActionType:
		return &InitAction{UserID: req.UserID}, nil
	}

	return nil, fmt.Errorf("invalid game action type %s", req.Type)
}
