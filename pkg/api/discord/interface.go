package discord

import "context"

type IEndpoint interface {
	WithUser(string) IEndpoint
	HasAddedBot(ctx context.Context, guildID string) (bool, error)
	CheckMember(ctx context.Context, guildID string) (bool, error)
	GetGuildFromCode(ctx context.Context, code string) (Guild, error)
}
