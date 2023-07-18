package domain

import (
	"context"

	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/domain/notification/event"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type ChatDomain interface {
	CreateMessage(context.Context, *model.CreateMessageRequest) (*model.CreateMessageResponse, error)
}

type chatDomain struct {
	notificationEngineCaller client.NotificationEngineCaller
}

func NewChatDomain(notificationEngineCaller client.NotificationEngineCaller) *chatDomain {
	return &chatDomain{notificationEngineCaller: notificationEngineCaller}
}

func (d *chatDomain) CreateMessage(
	ctx context.Context, req *model.CreateMessageRequest,
) (*model.CreateMessageResponse, error) {
	err := d.notificationEngineCaller.Emit(ctx, event.New(
		&event.MessageCreatedEvent{},
		event.Metadata{
			To: req.CommunityID,
		},
	))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot emit message: %v", err)
		return nil, errorx.Unknown
	}

	return &model.CreateMessageResponse{}, nil
}
