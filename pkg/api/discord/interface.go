package discord

import "context"

type IEndpoint interface {
	GetMe(ctx context.Context, token string) (User, error)
	HasAddedBot(ctx context.Context, guildID string) (bool, error)
	GetMember(ctx context.Context, guildID, userID string) (Member, error)
	CheckCode(ctx context.Context, guildID, code string) error
	GetCode(ctx context.Context, guildID, code string) (InviteCode, error)
	GetGuild(ctx context.Context, guildID string) (Guild, error)
	GetRoles(ctx context.Context, guildID string) ([]Role, error)
	GiveRole(ctx context.Context, guildID, userID, roleID string) error
}
