package domain

import (
	"context"

	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/domain/notification/event"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type ChatDomain interface {
	CreateMessage(context.Context, *model.CreateMessageRequest) (*model.CreateMessageResponse, error)
}

type chatDomain struct {
	chatMessageRepo repository.ChatMessageRepository

	notificationEngineCaller client.NotificationEngineCaller
}

func NewChatDomain(
	chatMessageRepo repository.ChatMessageRepository,
	notificationEngineCaller client.NotificationEngineCaller,
) *chatDomain {
	return &chatDomain{
		chatMessageRepo:          chatMessageRepo,
		notificationEngineCaller: notificationEngineCaller,
	}
}

func (d *chatDomain) CreateMessage(
	ctx context.Context, req *model.CreateMessageRequest,
) (*model.CreateMessageResponse, error) {
	messageID := xcontext.SnowFlake(ctx).Generate().Int64()
	msg := entity.ChatMessage{
		ID:          messageID,
		AuthorID:    xcontext.RequestUserID(ctx),
		ChannelID:   req.ChannelID,
		Content:     req.Content,
		Attachments: req.Attachments,
	}
	err := d.chatMessageRepo.Create(ctx, &msg)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot insert message to database: %v", err)
		return nil, errorx.Unknown
	}

	// TODO: Convert channelID to communityID, then put to metadata of Emit() method.
	msgEvent := &event.MessageCreatedEvent{Message: convertChatMessage(&msg, nil)}
	err = d.notificationEngineCaller.Emit(ctx, event.New(msgEvent, nil))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot emit message event: %v", err)
		return nil, errorx.Unknown
	}

	return &model.CreateMessageResponse{ID: messageID}, nil
}
