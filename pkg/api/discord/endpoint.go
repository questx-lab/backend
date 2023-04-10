package discord

import (
	"context"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/pkg/api"
)

const apiURL = "https://discord.com/api"

const userAgent = "DiscordBot (https://questx.com, 1.0)"

type Endpoint struct {
	BotToken string
	BotID    string

	UserID string
}

func New(ctx context.Context, cfg config.DiscordConfigs) *Endpoint {
	return &Endpoint{
		BotToken: cfg.BotToken,
		BotID:    cfg.BotID,
	}
}

func (e *Endpoint) WithUser(id string) IEndpoint {
	clone := *e
	clone.UserID = id
	return &clone
}

func (e *Endpoint) HasAddedBot(ctx context.Context, guildID string) (bool, error) {
	resp, err := api.New(apiURL, "/guilds/%s/members/%s", guildID, e.BotID).
		GET(ctx, api.OAuth2("Bot", e.BotToken), api.UserAgent(userAgent))
	if err != nil {
		return false, err
	}

	// If response has the field of code, an error is returned.
	if _, err := resp.GetInt("code"); err == nil {
		return false, nil
	}

	return true, nil
}

func (e *Endpoint) CheckMember(ctx context.Context, guildID string) (bool, error) {
	resp, err := api.New(apiURL, "/guilds/%s/members/%s", guildID, e.UserID).
		GET(ctx, api.OAuth2("Bot", e.BotToken), api.UserAgent(userAgent))
	if err != nil {
		return false, err
	}

	// If response has the field of code, an error is returned.
	if _, err := resp.GetInt("code"); err == nil {
		return false, nil
	}

	return true, nil
}

func (e *Endpoint) GetGuildFromCode(ctx context.Context, code string) (Guild, error) {
	resp, err := api.New(apiURL, "/invites/%s", code).
		GET(ctx, api.OAuth2("Bot", e.BotToken), api.UserAgent(userAgent))
	if err != nil {
		return Guild{}, err
	}

	id, err := resp.GetString("id")
	if err != nil {
		return Guild{}, err
	}

	return Guild{ID: id}, nil
}
