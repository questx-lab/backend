package discord

import "context"

type IEndpoint interface {
	WithUser(string) IEndpoint
	GetMe(ctx context.Context, token string) (User, error)
	HasAddedBot(ctx context.Context, guildID string) (bool, error)
	CheckMember(ctx context.Context, guildID string) (bool, error)
	CheckCode(ctx context.Context, guildID, code string) error
	GetGuild(ctx context.Context, guildID string) (Guild, error)
	GetRoles(ctx context.Context, guildID string) ([]Role, error)
	GiveRole(ctx context.Context, guildID, roleID string) error
}
