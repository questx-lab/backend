package gameengine

import (
	"context"
	"encoding/json"
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
	case *ChangeCharacterAction:
		resp.Value = map[string]any{
			"character": t.userCharacter,
		}

	case *CreateCharacterAction:
		resp.Value = map[string]any{
			"character": t.Character,
		}

	case *BuyCharacterAction:
		// No value.

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
			OwnerID:   req.UserID,
			Direction: directionEnum,
			Position:  Position{int(x), int(y)},
		}, nil

	case JoinAction{}.Type():
		return &JoinAction{OwnerID: req.UserID}, nil

	case ExitAction{}.Type():
		return &ExitAction{OwnerID: req.UserID}, nil

	case InitAction{}.Type():
		return &InitAction{OwnerID: req.UserID}, nil

	case MessageAction{}.Type():
		msg, ok := req.Value["message"].(string)
		if !ok {
			return nil, errors.New("message must be a string")
		}

		return &MessageAction{
			OwnerID:   req.UserID,
			Message:   msg,
			CreatedAt: time.Now(),
		}, nil

	case EmojiAction{}.Type():
		emoji, ok := req.Value["emoji"].(string)
		if !ok {
			return nil, errors.New("emoji must be a string")
		}

		return &EmojiAction{
			OwnerID: req.UserID,
			Emoji:   emoji,
		}, nil

	case StartLuckyboxEventAction{}.Type():
		eventID, ok := req.Value["event_id"].(string)
		if !ok {
			return nil, errors.New("event_id must be a string")
		}

		return &StartLuckyboxEventAction{
			OwnerID: req.UserID,
			EventID: eventID,
		}, nil

	case StopLuckyboxEventAction{}.Type():
		eventID, ok := req.Value["event_id"].(string)
		if !ok {
			return nil, errors.New("event_id must be a string")
		}

		return &StopLuckyboxEventAction{
			OwnerID: req.UserID,
			EventID: eventID,
		}, nil

	case CollectLuckyboxAction{}.Type():
		luckyboxID, ok := req.Value["luckybox_id"].(string)
		if !ok {
			return nil, errors.New("luckybox_id must be a string")
		}

		return &CollectLuckyboxAction{
			OwnerID:    req.UserID,
			LuckyboxID: luckyboxID,
		}, nil

	case CleanupProxyAction{}.Type():
		liveProxyIDs, ok := req.Value["live_proxy_ids"].([]string)
		if !ok {
			return nil, errors.New("live_proxy_ids must be an array of string")
		}

		return &CleanupProxyAction{OwnerID: req.UserID, LiveProxyIDs: liveProxyIDs}, nil
	case ChangeCharacterAction{}.Type():
		characterID, ok := req.Value["character_id"].(string)
		if !ok {
			return nil, errors.New("character_id must be a string")
		}

		return &ChangeCharacterAction{
			OwnerID:     req.UserID,
			CharacterID: characterID,
		}, nil

	case CreateCharacterAction{}.Type():
		b, err := json.Marshal(req.Value)
		if err != nil {
			return nil, err
		}

		var character Character
		if err := json.Unmarshal(b, &character); err != nil {
			return nil, err
		}

		return &CreateCharacterAction{
			OwnerID:   req.UserID,
			Character: character,
		}, nil

	case BuyCharacterAction{}.Type():
		characterID, ok := req.Value["character_id"].(string)
		if !ok {
			return nil, errors.New("character_id must be a string")
		}

		buyUserID, ok := req.Value["buy_user_id"].(string)
		if !ok {
			return nil, errors.New("buy_user_id must be a string")
		}
		return &BuyCharacterAction{
			OwnerID:     req.UserID,
			BuyUserID:   buyUserID,
			CharacterID: characterID,
		}, nil
	}

	return nil, fmt.Errorf("invalid game action type %s", req.Type)
}
