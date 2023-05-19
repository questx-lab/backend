package discord

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/puzpuzpuz/xsync"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/pkg/api"
)

const apiURL = "https://discord.com/api"
const userAgent = "DiscordBot (https://questx.com, 1.0)"
const iso8601 = "2006-01-02T15:04:05.000000+00:00"

const (
	giveRoleResource       = "give_role"
	getGuildInviteResource = "get_guild_invite"
)

type Endpoint struct {
	BotToken string
	BotID    string

	apiGenerator      api.Generator
	rateLimitResource *xsync.MapOf[string, *xsync.MapOf[string, time.Time]]
}

func New(cfg config.DiscordConfigs) *Endpoint {
	return &Endpoint{
		BotToken:          cfg.BotToken,
		BotID:             cfg.BotID,
		apiGenerator:      api.NewGenerator(),
		rateLimitResource: xsync.NewMapOf[*xsync.MapOf[string, time.Time]](),
	}
}

func (e *Endpoint) GetMe(ctx context.Context, token string) (User, error) {
	resp, err := e.apiGenerator.New(apiURL, "/users/@me").
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
	if err != nil {
		return User{}, err
	}

	return User{ID: id}, nil
}

func (e *Endpoint) HasAddedBot(ctx context.Context, guildID string) (bool, error) {
	resp, err := e.apiGenerator.New(apiURL, "/guilds/%s/members/%s", guildID, e.BotID).
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

func (e *Endpoint) CheckMember(ctx context.Context, guildID, userID string) (bool, error) {
	resp, err := e.apiGenerator.New(apiURL, "/guilds/%s/members/%s", guildID, userID).
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

func (e *Endpoint) CheckCode(ctx context.Context, guildID, code string) error {
	if err := e.checkLimitingResource(getGuildInviteResource, guildID); err != nil {
		return err
	}

	resp, err := e.apiGenerator.New(apiURL, "/guilds/%s/invites", guildID).
		Header("User-Agent", userAgent).
		GET(ctx, api.OAuth2("Bot", e.BotToken))
	if err != nil {
		return err
	}

	if err := e.checkTooManyRequest(resp, getGuildInviteResource, guildID); err != nil {
		return err
	}

	array, ok := resp.Body.(api.Array)
	if !ok {
		return errors.New("invalid response")
	}

	for _, obj := range array {
		c, err := obj.GetString("code")
		if err != nil {
			return err
		}

		if c != code {
			continue
		}

		maxUses, err := obj.GetInt("max_uses")
		if err != nil {
			return err
		}

		uses, err := obj.GetInt("uses")
		if err != nil {
			return err
		}

		if uses >= maxUses && maxUses != 0 {
			// If this link is nolimit uses, maxUses == 0.
			continue
		}

		createdAt, err := obj.GetTime("created_at", iso8601)
		if err != nil {
			return err
		}

		maxAge, err := obj.GetInt("max_age")
		if err != nil {
			return err
		}

		maxAgeDuration := time.Second * time.Duration(maxAge)
		if createdAt.Add(maxAgeDuration).After(time.Now()) {
			return nil
		}
	}

	return errors.New("invalid code")
}

func (e *Endpoint) GetCode(ctx context.Context, guildID, code string) (InviteCode, error) {
	if err := e.checkLimitingResource(getGuildInviteResource, guildID); err != nil {
		return InviteCode{}, err
	}

	resp, err := e.apiGenerator.New(apiURL, "/guilds/%s/invites", guildID).
		Header("User-Agent", userAgent).
		GET(ctx, api.OAuth2("Bot", e.BotToken))
	if err != nil {
		return InviteCode{}, err
	}

	if err := e.checkTooManyRequest(resp, getGuildInviteResource, guildID); err != nil {
		return InviteCode{}, err
	}

	array, ok := resp.Body.(api.Array)
	if !ok {
		return InviteCode{}, errors.New("invalid response")
	}

	for _, obj := range array {
		c, err := obj.GetString("code")
		if err != nil {
			return InviteCode{}, err
		}

		if c != code {
			continue
		}

		maxUses, err := obj.GetInt("max_uses")
		if err != nil {
			return InviteCode{}, err
		}

		uses, err := obj.GetInt("uses")
		if err != nil {
			return InviteCode{}, err
		}

		createdAt, err := obj.GetTime("created_at", iso8601)
		if err != nil {
			return InviteCode{}, err
		}

		maxAge, err := obj.GetInt("max_age")
		if err != nil {
			return InviteCode{}, err
		}

		inviterID, err := obj.GetString("inviter.id")
		if err != nil {
			return InviteCode{}, err
		}

		return InviteCode{
			Code:      c,
			Uses:      uses,
			MaxUses:   maxUses,
			MaxAge:    time.Second * time.Duration(maxAge),
			CreatedAt: createdAt,
			Inviter:   User{ID: inviterID},
		}, nil

	}

	return InviteCode{}, errors.New("invalid code")
}

func (e *Endpoint) GetRoles(ctx context.Context, guildID string) ([]Role, error) {
	resp, err := e.apiGenerator.New(apiURL, "/guilds/%s/roles", guildID).
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
	resp, err := e.apiGenerator.New(apiURL, "/guilds/%s", guildID).
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

func (e *Endpoint) GiveRole(ctx context.Context, guildID, userID, roleID string) error {
	if err := e.checkLimitingResource(giveRoleResource, guildID); err != nil {
		return err
	}

	resp, err := e.apiGenerator.New(apiURL, "/guilds/%s/members/%s/roles/%s", guildID, userID, roleID).
		Header("User-Agent", userAgent).
		PUT(ctx, api.OAuth2("Bot", e.BotToken))
	if err != nil {
		return err
	}

	if err := e.checkTooManyRequest(resp, giveRoleResource, guildID); err != nil {
		return err
	}

	return nil
}

func (e *Endpoint) checkLimitingResource(resource, identifier string) error {
	if limit, ok := e.rateLimitResource.Load(resource); ok {
		if resetAt, ok := limit.Load(identifier); ok {
			if resetAt.After(time.Now()) {
				return wrapRateLimit(resetAt.Unix())
			}

			// If the rate limit is reset, delete the limit for this resource.
			limit.Delete(identifier)
		}
	}

	return nil
}

func (e *Endpoint) checkTooManyRequest(resp *api.Response, resource, identifier string) error {
	if resp.Code == http.StatusTooManyRequests {
		resetAt, err := strconv.Atoi(resp.Header.Get("X-Ratelimit-Reset"))
		if err != nil {
			return err
		}

		resourceLimiter, _ := e.rateLimitResource.LoadOrStore(resource, xsync.NewMapOf[time.Time]())
		resourceLimiter.Store(identifier, time.Unix(int64(resetAt), 0))
		return wrapRateLimit(int64(resetAt))
	}

	return nil
}
