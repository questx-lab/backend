package telegram

import "context"

type IEndpoint interface {
	GetChat(ctx context.Context, chatID string) (Chat, error)
	GetMember(ctx context.Context, chatID, userID string) (User, error)
}
