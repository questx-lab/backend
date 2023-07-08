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
	Apply(ctx context.Context, proxyID string, g *GameState) ([]model.GameActionServerRequest, error)
}

func formatAction(a Action) (model.GameActionServerResponse, error) {
	resp := model.GameActionServerResponse{
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
			"player":    t.user.Character, // TODO: Not modify key for back-compatible.
			"user":      t.user.User,
			"position":  t.user.PixelPosition.TopLeftToCenter(t.user.Character.Size),
			"direction": t.user.Direction,
		}

	case *ExitAction:
		// No value.

	case *InitAction:
		resp.Value = map[string]any{
			"users":           t.initialUsers,
			"message_history": t.messageHistory,
			"luckyboxes":      t.luckyboxes,
		}

	case *MessageAction:
		resp.Value = map[string]any{
			"user":       t.user,
			"message":    t.Message,
			"created_at": t.CreatedAt.Format(time.RFC3339Nano),
		}

	case *EmojiAction:
		resp.Value = map[string]any{
			"emoji": t.Emoji,
		}

	case *StartLuckyboxEventAction:
		resp.Value = map[string]any{
			"luckyboxes": t.newLuckyboxes,
		}

	case *StopLuckyboxEventAction:
		resp.Value = map[string]any{
			"luckyboxes": t.removedLuckyboxes,
		}

	case *CollectLuckyboxAction:
		resp.Value = map[string]any{
			"luckybox": t.luckybox,
		}

	case *CleanupProxyAction:
		// No Value

	default:
		return model.GameActionServerResponse{}, fmt.Errorf("not set up action %T", a)
	}

	return resp, nil
}

func parseAction(req model.GameActionServerRequest) (Action, error) {
	switch req.Type {
	case MoveAction{}.Type():
		direction, ok := req.Value["direction"].(string)
		if !ok {
			return nil, errors.New("direction must be a string")
		}

		directionEnum, err := enum.ToEnum[entity.DirectionType](direction)
		if err != nil {
			return nil, err
		}

		x, ok := req.Value["x"].(float64)
		if !ok {
			return nil, errors.New("x must be a number")
		}

		y, ok := req.Value["y"].(float64)
		if !ok {
			return nil, errors.New("y must be a number")
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
		msg, ok := req.Value["message"].(string)
		if !ok {
			return nil, errors.New("message must be a string")
		}

		return &MessageAction{
			UserID:    req.UserID,
			Message:   msg,
			CreatedAt: time.Now(),
		}, nil

	case EmojiAction{}.Type():
		emoji, ok := req.Value["emoji"].(string)
		if !ok {
			return nil, errors.New("emoji must be a string")
		}

		return &EmojiAction{
			UserID: req.UserID,
			Emoji:  emoji,
		}, nil

	case StartLuckyboxEventAction{}.Type():
		eventID, ok := req.Value["event_id"].(string)
		if !ok {
			return nil, errors.New("event_id must be a string")
		}

		return &StartLuckyboxEventAction{
			UserID:  req.UserID,
			EventID: eventID,
		}, nil

	case StopLuckyboxEventAction{}.Type():
		eventID, ok := req.Value["event_id"].(string)
		if !ok {
			return nil, errors.New("event_id must be a string")
		}

		return &StopLuckyboxEventAction{
			UserID:  req.UserID,
			EventID: eventID,
		}, nil

	case CollectLuckyboxAction{}.Type():
		luckyboxID, ok := req.Value["luckybox_id"].(string)
		if !ok {
			return nil, errors.New("luckybox_id must be a string")
		}

		return &CollectLuckyboxAction{
			UserID:     req.UserID,
			LuckyboxID: luckyboxID,
		}, nil

	case CleanupProxyAction{}.Type():
		liveProxyIDs, ok := req.Value["live_proxy_ids"].([]string)
		if !ok {
			return nil, errors.New("live_proxy_ids must be an array of string")
		}

		return &CleanupProxyAction{UserID: req.UserID, LiveProxyIDs: liveProxyIDs}, nil
	}

	return nil, fmt.Errorf("invalid game action type %s", req.Type)
}
