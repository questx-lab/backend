package domain

import (
	"context"
	"errors"

	"github.com/gocql/gocql"
	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/notification/event"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/numberutil"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type ChatDomain interface {
	GetList(context.Context, *model.GetListMessageRequest) (*model.GetListMessageResponse, error)
	CreateChannel(context.Context, *model.CreateChannelRequest) (*model.CreateChannelResponse, error)
	CreateMessage(context.Context, *model.CreateMessageRequest) (*model.CreateMessageResponse, error)
	AddReaction(context.Context, *model.AddReactionRequest) (*model.AddReactionResponse, error)
}

type chatDomain struct {
	communityRepo             repository.CommunityRepository
	chatMessageRepo           repository.ChatMessageRepository
	chatChannelRepo           repository.ChatChannelRepository
	chatReactionRepo          repository.ChatReactionRepository
	chatReactionStatisticRepo repository.ChatReactionStatisticRepository
	chatChannelBucketRepo     repository.ChatChannelBucketRepository

	roleVerifier             *common.CommunityRoleVerifier
	notificationEngineCaller client.NotificationEngineCaller
}

func NewChatDomain(
	communityRepo repository.CommunityRepository,
	chatMessageRepo repository.ChatMessageRepository,
	chatChannelRepo repository.ChatChannelRepository,
	chatReactionRepo repository.ChatReactionRepository,
	chatReactionStatistic repository.ChatReactionStatisticRepository,
	chatChannelBucketRepo repository.ChatChannelBucketRepository,
	notificationEngineCaller client.NotificationEngineCaller,
	roleVerifier *common.CommunityRoleVerifier,
) *chatDomain {
	return &chatDomain{
		communityRepo:             communityRepo,
		chatMessageRepo:           chatMessageRepo,
		chatChannelRepo:           chatChannelRepo,
		chatReactionRepo:          chatReactionRepo,
		chatReactionStatisticRepo: chatReactionStatistic,
		roleVerifier:              roleVerifier,
		chatChannelBucketRepo:     chatChannelBucketRepo,
		notificationEngineCaller:  notificationEngineCaller,
	}
}

func (d *chatDomain) CreateChannel(
	ctx context.Context, req *model.CreateChannelRequest,
) (*model.CreateChannelResponse, error) {
	if req.CommunityHandle == "" {
		return nil, errorx.New(errorx.BadRequest, "Require community handle")
	}

	if req.ChannelName == "" {
		return nil, errorx.New(errorx.BadRequest, "Require channel name")
	}

	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.roleVerifier.Verify(ctx, community.ID); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	count, err := d.chatChannelRepo.CountByCommunityID(ctx, community.ID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot count the number of chanels in community: %v", err)
		return nil, errorx.Unknown
	}

	if count >= 10 {
		return nil, errorx.New(errorx.Unavailable, "Your community had too many channels")
	}

	channel := &entity.ChatChannel{
		SnowFlakeBase: entity.SnowFlakeBase{ID: xcontext.SnowFlake(ctx).Generate().Int64()},
		CommunityID:   community.ID,
		Name:          req.ChannelName,
		LastMessageID: 0,
	}

	if err := d.chatChannelRepo.Create(ctx, channel); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create channel: %v", err)
		return nil, errorx.Unknown
	}

	ev := event.New(
		&event.ChannelCreatedEvent{ChatChannel: convertChatChannel(channel)},
		&event.Metadata{To: channel.CommunityID},
	)
	if err := d.notificationEngineCaller.Emit(ctx, ev); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot emit channel event: %v", err)
		return nil, errorx.Unknown
	}

	return &model.CreateChannelResponse{ID: channel.ID}, nil
}

func (d *chatDomain) CreateMessage(
	ctx context.Context, req *model.CreateMessageRequest,
) (*model.CreateMessageResponse, error) {
	if req.Content == "" && len(req.Attachments) == 0 {
		return nil, errorx.New(errorx.BadRequest, "Require content or attachments")
	}

	channel, err := d.chatChannelRepo.GetByID(ctx, req.ChannelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found channel")
		}

		xcontext.Logger(ctx).Errorf("Cannot get channel: %v", err)
		return nil, errorx.Unknown
	}
	id := xcontext.SnowFlake(ctx).Generate().Int64()
	msg := entity.ChatMessage{
		ID:          id,
		Bucket:      numberutil.BucketFrom(id),
		AuthorID:    xcontext.RequestUserID(ctx),
		ChannelID:   req.ChannelID,
		Content:     req.Content,
		Attachments: req.Attachments,
	}

	if err := d.chatMessageRepo.Create(ctx, &msg); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create message: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.chatChannelBucketRepo.Increment(ctx, msg.ChannelID, msg.Bucket); err != nil {
		xcontext.Logger(ctx).Errorf("Unable to increase channel bucket: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.chatChannelRepo.UpdateLastMessageByID(ctx, channel.ID, msg.ID); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update last message of channel: %v", err)
		return nil, errorx.Unknown
	}

	ev := event.New(
		&event.MessageCreatedEvent{ChatMessage: convertChatMessage(&msg, nil)},
		&event.Metadata{To: channel.CommunityID},
	)
	if err := d.notificationEngineCaller.Emit(ctx, ev); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot emit message event: %v", err)
		return nil, errorx.Unknown
	}

	return &model.CreateMessageResponse{ID: msg.ID}, nil
}

func (d *chatDomain) AddReaction(
	ctx context.Context, req *model.AddReactionRequest,
) (*model.AddReactionResponse, error) {
	_, err := d.chatMessageRepo.Get(ctx, req.MessageID, req.ChannelID)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found message")
		}

		xcontext.Logger(ctx).Errorf("Cannot get message: %v", err)
		return nil, errorx.Unknown
	}

	channel, err := d.chatChannelRepo.GetByID(ctx, req.ChannelID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get message: %v", err)
		return nil, errorx.Unknown
	}

	_, err = d.chatReactionRepo.Get(ctx, xcontext.RequestUserID(ctx), req.MessageID, req.Emoji)
	if err != nil && !errors.Is(err, gocql.ErrNotFound) {
		xcontext.Logger(ctx).Errorf("Cannot get existing reaction record: %v", err)
		return nil, errorx.Unknown
	}

	if err == nil {
		return nil, errorx.New(errorx.Unavailable, "Cannot reaction an emoji for twice")
	}

	err = d.chatReactionRepo.Create(ctx, &entity.ChatReaction{
		MessageID: req.MessageID,
		UserID:    xcontext.RequestUserID(ctx),
		Emoji:     req.Emoji,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create reaction: %v", err)
		return nil, errorx.Unknown
	}

	err = d.chatReactionStatisticRepo.IncreaseCount(ctx, req.MessageID, req.Emoji)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot increase reaction count: %v", err)
		return nil, errorx.Unknown
	}

	ev := event.New(
		&event.ReactionAddedEvent{
			MessageID: req.MessageID,
			UserID:    xcontext.RequestUserID(ctx),
			Emoji:     req.Emoji,
		},
		&event.Metadata{To: channel.CommunityID},
	)
	if err := d.notificationEngineCaller.Emit(ctx, ev); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot emit add reaction event: %v", err)
		return nil, errorx.Unknown
	}

	return &model.AddReactionResponse{}, nil
}

func (d *chatDomain) GetList(
	ctx context.Context, req *model.GetListMessageRequest,
) (*model.GetListMessageResponse, error) {
	if req.Limit > 50 {
		return nil, errorx.New(errorx.BadRequest, "Maximum of limit is 50")
	}

	messages, err := d.chatMessageRepo.GetListByLastMessage(ctx, repository.LastMessageFilter{
		ChannelID:     req.ChannelID,
		LastMessageID: req.LastMessageID,
		Limit:         req.Limit,
		FromBucket:    numberutil.BucketFrom(req.LastMessageID),
		ToBucket:      numberutil.BucketFrom(req.ChannelID),
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Unable to get list message: %v", err)
		return nil, errorx.Unknown
	}

	messageIDs := []int64{}
	for _, mess := range messages {
		messageIDs = append(messageIDs, mess.ID)
	}

	reactions, err := d.chatReactionStatisticRepo.GetListByMessages(ctx, messageIDs)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Unable to get list reaction message: %v", err)
		return nil, errorx.Unknown
	}

	myID := xcontext.RequestUserID(ctx)
	reactionStates := make(map[int64][]model.ChatReactionState)
	for _, reaction := range reactions {
		_, myReactionErr := d.chatReactionRepo.Get(ctx, myID, reaction.MessageID, reaction.Emoji)
		if myReactionErr != nil && !errors.Is(myReactionErr, gocql.ErrNotFound) {
			xcontext.Logger(ctx).Errorf("Cannot get reaction: %v", myReactionErr)
			return nil, errorx.Unknown
		}

		reactionStates[reaction.MessageID] = append(reactionStates[reaction.MessageID], model.ChatReactionState{
			Emoji: reaction.Emoji,
			Count: reaction.Count,
			Me:    myReactionErr == nil,
		})
	}

	var msgResp []model.ChatMessage
	for _, msg := range messages {
		msgResp = append(msgResp, convertChatMessage(&msg, reactionStates[msg.ID]))
	}

	return &model.GetListMessageResponse{Messages: msgResp}, nil
}
