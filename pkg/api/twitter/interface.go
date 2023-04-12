package twitter

import "context"

type IEndpoint interface {
	WithUser(string) IEndpoint
	GetUser(context.Context, string) (User, error)
	CheckFollowing(context.Context, string) (bool, error)
}
