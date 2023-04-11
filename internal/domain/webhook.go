package domain

import (
	"crypto/ed25519"
	"encoding/hex"
	"net/http"

	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/discord"
	"github.com/questx-lab/backend/pkg/errorx"
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
	pubKeyStr := "693acd285950f5de143d85ce2afd0f43d1f8f3a5cb90321da71dc07d24962d61"
	pubKey, err := hex.DecodeString(pubKeyStr)
	if err != nil {
		return nil, err
	}

	if err := discord.Verify(ctx.Request(), ed25519.PublicKey(pubKey)); err != nil {
		ctx.Writer().Header().Set("Content-Type", "application/json")
		ctx.Writer().WriteHeader(http.StatusUnauthorized)
		return nil, errorx.New(errorx.Unauthenticated, "signature mismatch")
	}
	return &model.PostDiscordInteractResponse{
		Type: 1,
		Data: "hello world",
	}, nil
}
