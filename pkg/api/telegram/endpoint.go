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

func New(ctx context.Context, cfg config.TelegramConfigs) *Endpoint {
	return &Endpoint{
		BotToken:     cfg.BotToken,
		apiGenerator: api.NewGenerator(),
	}
}

func (e *Endpoint) GetAdministrators(ctx context.Context, chatID string) ([]User, error) {
	resp, err := e.apiGenerator.New(apiURL, "/bot%s/getChatAdministrators", e.BotToken).
		Query(api.Parameter{"chat_id": chatID}).
		GET(ctx)
	if err != nil {
		return nil, err
	}

	body, ok := resp.Body.(api.JSON)
	if !ok {
		return nil, errors.New("invalid body type")
	}

	if ok, err := body.GetBool("ok"); err != nil || !ok {
		return nil, fmt.Errorf("invalid response")
	}

	results, err := body.GetArray("result")
	if err != nil {
		return nil, err
	}

	var users []User
	for _, admin := range results {
		var canInvite bool
		if status, err := admin.Get("status"); err == nil && status == "creator" {
			canInvite = true
		} else if b, err := admin.GetBool("can_invite_users"); err == nil && b {
			canInvite = true
		}

		if !canInvite {
			continue
		}

		usr, err := admin.GetJSON("user")
		if err != nil {
			return nil, err
		}

		userID, err := usr.GetInt("id")
		if err != nil {
			return nil, err
		}

		users = append(users, User{ID: strconv.Itoa(userID)})
	}

	return users, nil
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
