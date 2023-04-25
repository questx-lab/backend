package discord

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/pkg/api"
)

const apiURL = "https://discord.com/api"
const userAgent = "DiscordBot (https://questx.com, 1.0)"
const iso8601 = "2006-01-02T15:04:05-0700"

var (
	giveRoleResource      = "give_role"
	getGuildInvteResource = "get_guild_invite"
)

type Endpoint struct {
	BotToken string
	BotID    string

	UserID string

	rateLimitResource map[string]map[string]time.Time
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

func (e *Endpoint) GetMe(ctx context.Context, token string) (User, error) {
	resp, err := api.New(apiURL, "/users/@me").
		Header("User-Agent", userAgent).
		GET(ctx, api.OAuth2("Bearer", token))
	if err != nil {
		return User{}, err
	}

	body, ok := resp.Body.(api.JSON)
	if !ok {
		return User{}, errors.New("invalid response")
	}

	// If response has the field of code, an error is returned.
	id, err := body.GetString("id")
	if err == nil {
		return User{}, nil
	}

	return User{ID: id}, nil
}

func (e *Endpoint) HasAddedBot(ctx context.Context, guildID string) (bool, error) {
	resp, err := api.New(apiURL, "/guilds/%s/members/%s", guildID, e.BotID).
		Header("User-Agent", userAgent).
		GET(ctx, api.OAuth2("Bot", e.BotToken))
	if err != nil {
		return false, err
	}

	body, ok := resp.Body.(api.JSON)
	if !ok {
		return false, errors.New("invalid response")
	}

	// If response has the field of code, an error is returned.
	if _, err := body.GetInt("code"); err == nil {
		return false, nil
	}

	return true, nil
}

func (e *Endpoint) CheckMember(ctx context.Context, guildID string) (bool, error) {
	resp, err := api.New(apiURL, "/guilds/%s/members/%s", guildID, e.UserID).
		Header("User-Agent", userAgent).
		GET(ctx, api.OAuth2("Bot", e.BotToken))
	if err != nil {
		return false, err
	}

	body, ok := resp.Body.(api.JSON)
	if !ok {
		return false, errors.New("invalid response")
	}

	// If response has the field of code, an error is returned.
	if _, err := body.GetInt("code"); err == nil {
		return false, nil
	}

	return true, nil
}

func (e *Endpoint) GetCode(ctx context.Context, guildID string) (string, error) {
	if limit, ok := e.rateLimitResource[getGuildInvteResource]; ok {
		if resetAt, ok := limit[guildID]; ok {
			if resetAt.After(time.Now()) {
				return "", wrapRateLimit(resetAt.Unix())
			}

			// If the rate limit is reset, delete the limit for this resource.
			delete(limit, guildID)
		}
	}

	resp, err := api.New(apiURL, "/guilds/%s/invites", guildID).
		Header("User-Agent", userAgent).
		GET(ctx, api.OAuth2("Bot", e.BotToken))
	if err != nil {
		return "", err
	}

	if resp.Code == 429 {
		resetAt, err := strconv.Atoi(resp.Header.Get("X-Ratelimit-Reset"))
		if err != nil {
			return "", err
		}

		e.rateLimitResource[giveRoleResource][guildID] = time.Unix(int64(resetAt), 0)
		return "", wrapRateLimit(int64(resetAt))
	}

	array, ok := resp.Body.(api.Array)
	if !ok {
		return "", errors.New("invalid response")
	}

	for _, obj := range array {
		maxUses, err := obj.GetInt("max_uses")
		if err != nil {
			return "", err
		}

		uses, err := obj.GetInt("uses")
		if err != nil {
			return "", err
		}

		if uses >= maxUses {
			continue
		}

		code, err := obj.GetString("code")
		if err != nil {
			return "", err
		}

		created_at, err := obj.GetTime("created_at", iso8601)
		if err != nil {
			return "", err
		}

		maxAge, err := obj.GetInt("max_age")
		if err != nil {
			return "", err
		}

		maxAgeDuration := time.Second * time.Duration(maxAge)
		if created_at.Add(maxAgeDuration).Before(time.Now()) {
			return code, nil
		}
	}

	return "", errors.New("not found any suitable code")
}

func (e *Endpoint) GetRoles(ctx context.Context, guildID string) ([]Role, error) {
	resp, err := api.New(apiURL, "/guilds/%s/roles", guildID).
		Header("User-Agent", userAgent).
		GET(ctx, api.OAuth2("Bot", e.BotToken))
	if err != nil {
		return nil, err
	}

	array, ok := resp.Body.(api.Array)
	if !ok {
		return nil, errors.New("invalid response")
	}

	var roles []Role
	for _, role := range array {
		id, err := role.GetString("id")
		if err != nil {
			return nil, err
		}

		name, err := role.GetString("name")
		if err != nil {
			return nil, err
		}

		roles = append(roles, Role{ID: id, Name: name})
	}

	return roles, nil
}

func (e *Endpoint) GetGuild(ctx context.Context, guildID string) (Guild, error) {
	resp, err := api.New(apiURL, "/guilds/%s", guildID).
		Header("User-Agent", userAgent).
		GET(ctx, api.OAuth2("Bot", e.BotToken))
	if err != nil {
		return Guild{}, err
	}

	body, ok := resp.Body.(api.JSON)
	if !ok {
		return Guild{}, errors.New("invalid response")
	}

	id, err := body.GetString("id")
	if err != nil {
		return Guild{}, err
	}

	ownerID, err := body.GetString("owner_id")
	if err != nil {
		return Guild{}, err
	}

	return Guild{ID: id, OwnerID: ownerID}, nil
}

func (e *Endpoint) GiveRole(ctx context.Context, guildID, roleID string) error {
	if limit, ok := e.rateLimitResource[giveRoleResource]; ok {
		if resetAt, ok := limit[guildID]; ok {
			if resetAt.After(time.Now()) {
				return wrapRateLimit(resetAt.Unix())
			}

			// If the rate limit is reset, delete the limit for this resource.
			delete(limit, guildID)
		}
	}

	resp, err := api.New(apiURL, "/guilds/%s/members/%s/roles/%s", guildID, e.UserID, roleID).
		Header("User-Agent", userAgent).
		PUT(ctx, api.OAuth2("Bot", e.BotToken))
	if err != nil {
		return err
	}

	if resp.Code == 429 {
		resetAt, err := strconv.Atoi(resp.Header.Get("X-Ratelimit-Reset"))
		if err != nil {
			return err
		}

		e.rateLimitResource[giveRoleResource][guildID] = time.Unix(int64(resetAt), 0)
		return wrapRateLimit(int64(resetAt))
	}

	return nil
}
