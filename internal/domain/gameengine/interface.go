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

	case *ChangeCharacterAction:
		resp.Value = map[string]any{
			"character": t.character,
		}

	case *CreateCharacterAction:
		// No value.

	case *BuyCharacterAction:
		// No value.

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

		amount, ok := req.Value["amount"].(float64)
		if !ok {
			return nil, errors.New("amount must be a number")
		}

		pointPerBox, ok := req.Value["point_per_box"].(float64)
		if !ok {
			return nil, errors.New("point_per_box must be a number")
		}

		isRandom, ok := req.Value["is_random"].(bool)
		if !ok {
			return nil, errors.New("is_random must be a boolean")
		}

		return &StartLuckyboxEventAction{
			UserID:      req.UserID,
			EventID:     eventID,
			Amount:      int(amount),
			PointPerBox: int(pointPerBox),
			IsRandom:    isRandom,
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

	case ChangeCharacterAction{}.Type():
		characterID, ok := req.Value["character_id"].(string)
		if !ok {
			return nil, errors.New("character_id must be a string")
		}

		return &ChangeCharacterAction{
			UserID:      req.UserID,
			CharacterID: characterID,
		}, nil

	case CreateCharacterAction{}.Type():
		characterID, ok := req.Value["id"].(string)
		if !ok {
			return nil, errors.New("id must be a string")
		}

		characterName, ok := req.Value["name"].(string)
		if !ok {
			return nil, errors.New("name must be a string")
		}

		characterLevel, ok := req.Value["level"].(float64)
		if !ok {
			return nil, errors.New("level must be a number")
		}

		characterWidth, ok := req.Value["width"].(float64)
		if !ok {
			return nil, errors.New("width must be a number")
		}

		characterHeight, ok := req.Value["height"].(float64)
		if !ok {
			return nil, errors.New("height must be a number")
		}

		characterSpriteWidthRatio, ok := req.Value["sprite_width_ratio"].(float64)
		if !ok {
			return nil, errors.New("sprite_width_ratio must be a number")
		}

		characterSpriteHeightRatio, ok := req.Value["sprite_height_ratio"].(float64)
		if !ok {
			return nil, errors.New("sprite_height_ratio must be a number")
		}

		return &CreateCharacterAction{
			UserID:                     req.UserID,
			CharacterID:                characterID,
			CharacterName:              characterName,
			CharacterLevel:             int(characterLevel),
			CharacterWidth:             int(characterWidth),
			CharacterHeight:            int(characterHeight),
			CharacterSpriteWidthRatio:  characterSpriteWidthRatio,
			CharacterSpriteHeightRatio: characterSpriteHeightRatio,
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
			UserID:      req.UserID,
			BuyUserID:   buyUserID,
			CharacterID: characterID,
		}, nil
	}

	return nil, fmt.Errorf("invalid game action type %s", req.Type)
}
