package telegram

import "context"

type IEndpoint interface {
	GetAdministrators(ctx context.Context, chatID string) ([]User, error)
	GetMember(ctx context.Context, chatID, userID string) (User, error)
}
