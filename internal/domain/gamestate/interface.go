package gamestate

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
)

type Action interface {
	Apply(*GameState) error
}

func FormatAction(id int, a Action) (model.GameActionClientResponse, error) {
	switch t := a.(type) {
	case *MoveAction:
		return model.GameActionClientResponse{
			ID:     id,
			Type:   MoveActionType,
			UserID: t.UserID,
			Value:  map[string]any{"direction": t.Direction},
		}, nil

	case *JoinAction:
		return model.GameActionClientResponse{
			ID:     id,
			Type:   JoinActionType,
			UserID: t.UserID,
			Value: map[string]any{
				"position":  t.position,
				"direction": t.direction,
			},
		}, nil

	case *ExitAction:
		return model.GameActionClientResponse{
			ID:     id,
			Type:   ExitActionType,
			UserID: t.UserID,
			Value:  nil,
		}, nil

	default:
		return model.GameActionClientResponse{}, fmt.Errorf("not set up action %T", a)
	}
}

func FormatActionV2(id int, roomID string, a Action) (model.GameActionClientResponse, error) {
	switch t := a.(type) {
	case *Move:
		return model.GameActionClientResponse{
			ID:     id,
			Type:   moveActionType,
			UserID: t.UserID,
			Value:  map[string]any{"direction": t.Direction},
		}, nil

	default:
		return model.GameActionClientResponse{}, fmt.Errorf("not set up action %T", a)
	}
}

func ParseAction(req model.GameActionRouterRequest) (Action, error) {
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
	}

	return nil, fmt.Errorf("invalid game action type %s", req.Type)
}

func ClientActionToRouterAction(
	req model.GameActionClientRequest, roomID, userID string,
) model.GameActionRouterRequest {
	return model.GameActionRouterRequest{
		Type:   req.Type,
		Value:  req.Value,
		RoomID: roomID,
		UserID: userID,
	}
}
