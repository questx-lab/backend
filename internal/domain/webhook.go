package domain

import (
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type WebhookDomain interface {
	PostDiscordInteract(xcontext.Context, *model.PostDiscordInteractRequest) (*model.PostDiscordInteractResponse, error)
}

type webhookDomain struct {
}

func NewWebhookDomain() WebhookDomain {
	return &webhookDomain{}
}

func (d *webhookDomain) PostDiscordInteract(ctx xcontext.Context, req *model.PostDiscordInteractRequest) (*model.PostDiscordInteractResponse, error) {
	return &model.PostDiscordInteractResponse{}, nil
}
