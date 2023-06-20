package gameengine

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/enum"
)

type Action interface {
	// SendTo indicates clients to which the response of this action will be
	// sent. The returned array contains id of client need to receive the
	// response. Return nil to broadcast to all clients. Return an empty array
	// to not send to anyone.
	SendTo() []string

	// Type is the name of action. This should be a static string.
	Type() string

	// Owner is the client submit this action.
	Owner() string

	// Apply modifies game state based on the action.
	Apply(context.Context, *GameState) error
}

func formatAction(a Action) (model.GameActionResponse, error) {
	resp := model.GameActionResponse{
		UserID: a.Owner(),
		To:     a.SendTo(),
		Type:   a.Type(),
	}

	switch t := a.(type) {
	case *MoveAction:
		resp.Value = map[string]any{
			"direction": t.Direction,
			"x":         t.Position.X,
			"y":         t.Position.Y,
		}

	case *JoinAction:
		resp.Value = map[string]any{
			"player":    t.user.Player,
			"user":      t.user.User,
			"position":  t.user.PixelPosition.TopLeftToCenter(t.user.Player),
			"direction": t.user.Direction,
		}

	case *ExitAction:
		// No value.

	case *InitAction:
		resp.Value = map[string]any{
			"users":           t.initialUsers,
			"message_history": t.messageHistory,
		}

	case *MessageAction:
		resp.Value = map[string]any{
			"message":    t.Message,
			"created_at": t.CreatedAt.Format(time.RFC3339Nano),
		}

	default:
		return model.GameActionResponse{}, fmt.Errorf("not set up action %T", a)
	}

	return resp, nil
}

func parseAction(req model.GameActionServerRequest) (Action, error) {
	switch req.Type {
	case MoveAction{}.Type():
		direction, ok := req.Value["direction"].(string)
		if !ok {
			return nil, errors.New("invalid or not found direction")
		}

		directionEnum, err := enum.ToEnum[entity.DirectionType](direction)
		if err != nil {
			return nil, err
		}

		x, ok := req.Value["x"].(float64)
		if !ok {
			return nil, errors.New("invalid x")
		}

		y, ok := req.Value["y"].(float64)
		if !ok {
			return nil, errors.New("invalid y")
		}

		return &MoveAction{
			UserID:    req.UserID,
			Direction: directionEnum,
			Position:  Position{int(x), int(y)},
		}, nil

	case JoinAction{}.Type():
		return &JoinAction{UserID: req.UserID}, nil

	case ExitAction{}.Type():
		return &ExitAction{UserID: req.UserID}, nil

	case InitAction{}.Type():
		return &InitAction{UserID: req.UserID}, nil

	case MessageAction{}.Type():
		return &MessageAction{
			UserID:    req.UserID,
			Message:   req.Value["message"].(string),
			CreatedAt: time.Now(),
		}, nil
	}

	return nil, fmt.Errorf("invalid game action type %s", req.Type)
}
