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
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

type ChatDomain interface {
	GetMessages(context.Context, *model.GetMessagesRequest) (*model.GetMessagesResponse, error)
	GetChannles(context.Context, *model.GetChannelsRequest) (*model.GetChannelsResponse, error)
	CreateChannel(context.Context, *model.CreateChannelRequest) (*model.CreateChannelResponse, error)
	CreateMessage(context.Context, *model.CreateMessageRequest) (*model.CreateMessageResponse, error)
	DeleteMessage(context.Context, *model.DeleteMessageRequest) (*model.DeleteMessageResponse, error)
	AddReaction(context.Context, *model.AddReactionRequest) (*model.AddReactionResponse, error)
	RemoveReaction(context.Context, *model.RemoveReactionRequest) (*model.RemoveReactionResponse, error)
	GetUserReactions(context.Context, *model.GetUserReactionsRequest) (*model.GetUserReactionsResponse, error)
}

type chatDomain struct {
	communityRepo         repository.CommunityRepository
	chatMessageRepo       repository.ChatMessageRepository
	chatChannelRepo       repository.ChatChannelRepository
	chatReactionRepo      repository.ChatReactionRepository
	chatMemberRepo        repository.ChatMemberRepository
	chatChannelBucketRepo repository.ChatChannelBucketRepository
	userRepo              repository.UserRepository
	followerRepo          repository.FollowerRepository

	roleVerifier             *common.CommunityRoleVerifier
	notificationEngineCaller client.NotificationEngineCaller
}

func NewChatDomain(
	communityRepo repository.CommunityRepository,
	chatMessageRepo repository.ChatMessageRepository,
	chatChannelRepo repository.ChatChannelRepository,
	chatReactionRepo repository.ChatReactionRepository,
	chatMemberRepo repository.ChatMemberRepository,
	chatChannelBucketRepo repository.ChatChannelBucketRepository,
	userRepo repository.UserRepository,
	notificationEngineCaller client.NotificationEngineCaller,
	roleVerifier *common.CommunityRoleVerifier,
) *chatDomain {
	return &chatDomain{
		communityRepo:            communityRepo,
		chatMessageRepo:          chatMessageRepo,
		chatChannelRepo:          chatChannelRepo,
		chatReactionRepo:         chatReactionRepo,
		chatMemberRepo:           chatMemberRepo,
		chatChannelBucketRepo:    chatChannelBucketRepo,
		userRepo:                 userRepo,
		roleVerifier:             roleVerifier,
		notificationEngineCaller: notificationEngineCaller,
	}
}

func (d *chatDomain) GetChannles(
	ctx context.Context, req *model.GetChannelsRequest,
) (*model.GetChannelsResponse, error) {
	communityIDs := []string{}
	if req.CommunityHandle != "" {
		community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
		if err != nil {
			if errors.Is(err, gocql.ErrNotFound) {
				return nil, errorx.New(errorx.NotFound, "Not found community")
			}

			xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
			return nil, errorx.Unknown
		}

		communityIDs = append(communityIDs, community.ID)
	} else {
		followers, err := d.followerRepo.GetListByUserID(ctx, xcontext.RequestUserID(ctx))
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get followers: %v", err)
			return nil, errorx.Unknown
		}

		for _, f := range followers {
			communityIDs = append(communityIDs, f.CommunityID)
		}
	}

	channels, err := d.chatChannelRepo.GetByCommunityIDs(ctx, communityIDs)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get channels: %v", err)
		return nil, errorx.Unknown
	}

	clientChannels := []model.ChatChannel{}
	for _, c := range channels {
		clientChannels = append(clientChannels, convertChatChannel(&c, req.CommunityHandle))
	}

	return &model.GetChannelsResponse{Channels: clientChannels}, nil
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

	go func() {
		ev := event.New(
			&event.ChannelCreatedEvent{ChatChannel: convertChatChannel(channel, community.Handle)},
			&event.Metadata{ToCommunity: channel.CommunityID},
		)
		if err := d.notificationEngineCaller.Emit(ctx, ev); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot emit channel event: %v", err)
		}
	}()

	return &model.CreateChannelResponse{ID: channel.ID}, nil
}

func (d *chatDomain) CreateMessage(
	ctx context.Context, req *model.CreateMessageRequest,
) (*model.CreateMessageResponse, error) {
	if req.Content == "" && len(req.Attachments) == 0 {
		return nil, errorx.New(errorx.BadRequest, "Require content or attachments")
	}

	user, err := d.userRepo.GetByID(ctx, xcontext.RequestUserID(ctx))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get user information: %v", err)
		return nil, errorx.Unknown
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

	if err := d.chatChannelBucketRepo.Increase(ctx, msg.ChannelID, msg.Bucket); err != nil {
		xcontext.Logger(ctx).Errorf("Unable to increase channel bucket: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.chatChannelRepo.UpdateLastMessageByID(ctx, channel.ID, msg.ID); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update last message of channel: %v", err)
		return nil, errorx.Unknown
	}

	err = d.chatMemberRepo.Upsert(ctx, &entity.ChatMember{
		UserID:            xcontext.RequestUserID(ctx),
		ChannelID:         req.ChannelID,
		LastReadMessageID: msg.ID,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update last read message of member: %v", err)
		return nil, errorx.Unknown
	}

	go func() {
		ev := event.New(
			&event.MessageCreatedEvent{
				ChatMessage: convertChatMessage(&msg, convertUser(user, nil, false), nil),
			},
			&event.Metadata{ToCommunity: channel.CommunityID},
		)
		if err := d.notificationEngineCaller.Emit(ctx, ev); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot emit message event: %v", err)
		}
	}()
	return &model.CreateMessageResponse{ID: msg.ID}, nil
}

func (d *chatDomain) AddReaction(
	ctx context.Context, req *model.AddReactionRequest,
) (*model.AddReactionResponse, error) {
	if _, err := d.chatMessageRepo.Get(ctx, req.MessageID, req.ChannelID); err != nil {
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

	userID := xcontext.RequestUserID(ctx)
	isUserReacted, err := d.chatReactionRepo.CheckUserReaction(ctx, userID, req.MessageID, req.Emoji)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get existing reaction record: %v", err)
		return nil, errorx.Unknown
	}

	if isUserReacted {
		return nil, errorx.New(errorx.Unavailable, "Cannot reaction an emoji for twice")
	}

	if err := d.chatReactionRepo.Add(ctx, req.MessageID, req.Emoji, userID); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create reaction: %v", err)
		return nil, errorx.Unknown
	}

	go func() {
		ev := event.New(
			&event.ReactionAddedEvent{
				ChannelID: req.ChannelID,
				MessageID: req.MessageID,
				UserID:    xcontext.RequestUserID(ctx),
				Emoji:     req.Emoji,
			},
			&event.Metadata{ToCommunity: channel.CommunityID},
		)
		if err := d.notificationEngineCaller.Emit(ctx, ev); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot emit add reaction event: %v", err)
		}
	}()

	return &model.AddReactionResponse{}, nil
}

func (d *chatDomain) RemoveReaction(
	ctx context.Context, req *model.RemoveReactionRequest,
) (*model.RemoveReactionResponse, error) {
	if _, err := d.chatMessageRepo.Get(ctx, req.MessageID, req.ChannelID); err != nil {
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

	userID := xcontext.RequestUserID(ctx)
	isUserReacted, err := d.chatReactionRepo.CheckUserReaction(ctx, userID, req.MessageID, req.Emoji)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get existing reaction record: %v", err)
		return nil, errorx.Unknown
	}

	if !isUserReacted {
		return nil, errorx.New(errorx.Unavailable, "User has not react this emoji yet")
	}

	if err := d.chatReactionRepo.Remove(ctx, req.MessageID, req.Emoji, userID); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot remove reaction: %v", err)
		return nil, errorx.Unknown
	}

	go func() {
		ev := event.New(
			&event.ReactionRemovedEvent{
				ChannelID: req.ChannelID,
				MessageID: req.MessageID,
				UserID:    xcontext.RequestUserID(ctx),
				Emoji:     req.Emoji,
			},
			&event.Metadata{ToCommunity: channel.CommunityID},
		)
		if err := d.notificationEngineCaller.Emit(ctx, ev); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot emit add reaction event: %v", err)
		}
	}()

	return &model.RemoveReactionResponse{}, nil
}

func (d *chatDomain) GetMessages(
	ctx context.Context, req *model.GetMessagesRequest,
) (*model.GetMessagesResponse, error) {
	if req.Limit > 50 {
		return nil, errorx.New(errorx.BadRequest, "Maximum of limit is 50")
	}

	messages, err := d.chatMessageRepo.GetListByLastMessage(ctx, repository.LastMessageFilter{
		ChannelID:  req.ChannelID,
		Before:     req.Before,
		Limit:      req.Limit,
		FromBucket: numberutil.BucketFrom(req.Before),
		ToBucket:   numberutil.BucketFrom(req.ChannelID),
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Unable to get list message: %v", err)
		return nil, errorx.Unknown
	}

	messageIDs := []int64{}
	authorMap := map[string]entity.User{}
	for _, msg := range messages {
		messageIDs = append(messageIDs, msg.ID)
		authorMap[msg.AuthorID] = entity.User{}
	}

	authors, err := d.userRepo.GetByIDs(ctx, common.MapKeys(authorMap))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get authors information: %v", err)
		return nil, errorx.Unknown
	}

	for i := range authors {
		authorMap[authors[i].ID] = authors[i]
	}

	reactions, err := d.chatReactionRepo.GetByMessageIDs(ctx, messageIDs)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Unable to get list reaction message: %v", err)
		return nil, errorx.Unknown
	}

	myID := xcontext.RequestUserID(ctx)
	reactionStates := make(map[int64][]model.ChatReactionState)
	for _, reaction := range reactions {
		state := model.ChatReactionState{
			Emoji: reaction.Emoji,
			Count: len(reaction.UserIds),
		}
		if slices.Contains(reaction.UserIds, myID) {
			state.Me = true
		}

		reactionStates[reaction.MessageID] = append(reactionStates[reaction.MessageID], state)
	}

	msgResp := []model.ChatMessage{}
	for _, msg := range messages {
		author, ok := authorMap[msg.AuthorID]
		if !ok {
			xcontext.Logger(ctx).Errorf("Not found author info %s", msg.AuthorID)
			return nil, errorx.Unknown
		}

		msgResp = append(msgResp, convertChatMessage(
			&msg, convertUser(&author, nil, false), reactionStates[msg.ID]))
	}

	return &model.GetMessagesResponse{Messages: msgResp}, nil
}

func (d *chatDomain) GetUserReactions(ctx context.Context, req *model.GetUserReactionsRequest) (*model.GetUserReactionsResponse, error) {
	if req.Limit > 50 {
		return nil, errorx.New(errorx.BadRequest, "Maximum of limit is 50")
	}

	reaction, err := d.chatReactionRepo.Get(ctx, req.MessageID, req.Emoji)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Unable to get user reactions: %v", err)
		return nil, errorx.Unknown
	}

	if len(reaction.UserIds) == 0 {
		return &model.GetUserReactionsResponse{Users: []model.User{}}, nil
	}

	users, err := d.userRepo.GetByIDs(ctx, reaction.UserIds[:req.Limit])
	if err != nil {
		xcontext.Logger(ctx).Errorf("Unable to get users: %v", err)
		return nil, errorx.Unknown
	}

	respUsers := make([]model.User, 0, len(reaction.UserIds))
	for _, u := range users {
		respUsers = append(respUsers, convertUser(&u, nil, false))
	}

	return &model.GetUserReactionsResponse{Users: respUsers}, nil
}

func (d *chatDomain) DeleteMessage(ctx context.Context, req *model.DeleteMessageRequest) (*model.DeleteMessageResponse, error) {
	message, err := d.chatMessageRepo.Get(ctx, req.MessageID, req.ChannelID)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found message")
		}

		xcontext.Logger(ctx).Errorf("Cannot get message: %v", err)
		return nil, errorx.Unknown
	}

	channel, err := d.chatChannelRepo.GetByID(ctx, req.ChannelID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Unable to get channel: %v", err)
		return nil, errorx.Unknown
	}

	if message.AuthorID != xcontext.RequestUserID(ctx) {
		if err := d.roleVerifier.Verify(ctx, channel.CommunityID); err != nil {
			xcontext.Logger(ctx).Errorf("User doesn't have permission: %v", err)
			return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
		}
	}

	bucket := numberutil.BucketFrom(req.MessageID)

	if err := d.chatMessageRepo.Delete(ctx, req.ChannelID, bucket, req.MessageID); err != nil {
		xcontext.Logger(ctx).Errorf("Unable to delete message: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.chatChannelBucketRepo.Decrease(ctx, req.ChannelID, bucket); err != nil {
		xcontext.Logger(ctx).Errorf("Unable to decrease channel bucket: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.chatReactionRepo.RemoveByMessageID(ctx, req.MessageID); err != nil {
		xcontext.Logger(ctx).Errorf("Unable to remove chat reaction by message id: %v", err)
		return nil, errorx.Unknown
	}

	go func() {
		ev := event.New(
			&event.MessageDeletedEvent{MessageID: req.MessageID},
			&event.Metadata{ToCommunity: channel.CommunityID},
		)
		if err := d.notificationEngineCaller.Emit(ctx, ev); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot emit delete message event: %v", err)
		}
	}()

	return &model.DeleteMessageResponse{}, nil
}
