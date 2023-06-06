package telegram

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/pkg/api"
)

const apiURL = "https://api.telegram.org"

type Endpoint struct {
	BotToken string

	apiGenerator api.Generator
}

func New(cfg config.TelegramConfigs) *Endpoint {
	return &Endpoint{
		BotToken:     cfg.BotToken,
		apiGenerator: api.NewGenerator(),
	}
}

func (e *Endpoint) GetChat(ctx context.Context, chatID string) (Chat, error) {
	resp, err := e.apiGenerator.New(apiURL, "/bot%s/getChat", e.BotToken).
		Query(api.Parameter{"chat_id": chatID}).
		GET(ctx)
	if err != nil {
		return Chat{}, err
	}

	body, ok := resp.Body.(api.JSON)
	if !ok {
		return Chat{}, errors.New("invalid body type")
	}

	if ok, err := body.GetBool("ok"); err != nil || !ok {
		return Chat{}, fmt.Errorf("invalid response")
	}

	result, err := body.GetJSON("result")
	if err != nil {
		return Chat{}, err
	}

	id, err := result.GetInt("id")
	if err != nil {
		return Chat{}, err
	}

	chat := Chat{ID: id}

	_, err = result.Get("active_usernames")
	if err != nil {
		if !errors.Is(err, api.NotFoundKeyError) {
			return Chat{}, err
		}

		chat.IsPublic = false
	} else {
		chat.IsPublic = true
	}

	return chat, nil
}

func (e *Endpoint) GetMember(ctx context.Context, chatID, userID string) (User, error) {
	resp, err := e.apiGenerator.New(apiURL, "/bot%s/getChatMember", e.BotToken).
		Query(api.Parameter{
			"chat_id": chatID,
			"user_id": userID,
		}).
		GET(ctx)
	if err != nil {
		return User{}, err
	}

	body, ok := resp.Body.(api.JSON)
	if !ok {
		return User{}, errors.New("invalid body type")
	}

	if ok, err := body.GetBool("ok"); err != nil || !ok {
		return User{}, fmt.Errorf("invalid response")
	}

	result, err := body.GetJSON("result")
	if err != nil {
		return User{}, err
	}

	user, err := result.GetJSON("user")
	if err != nil {
		return User{}, err
	}

	id, err := user.GetInt("id")
	if err != nil {
		return User{}, err
	}

	return User{ID: strconv.Itoa(id)}, nil
}
