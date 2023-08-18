package domain

import (
	"context"
	"errors"
	"strings"

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
	"github.com/questx-lab/backend/pkg/xredis"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

// This indicates you need how many XP to reach to a level.
// NOTE: The below XP is specifying the current XP of user, not total XP.
var chatLevelConfigs = map[int]int{
	1: 5, 2: 6, 3: 7, 4: 8, 5: 9,
	6: 10, 7: 11, 8: 12, 9: 13, 10: 14,
	11: 15, 12: 20, 13: 25, 14: 30, 15: 35,
	16: 40, 17: 50, 18: 60, 19: 70, 20: 80,
	21: 90, 22: 100, 23: 110, 24: 120, 25: 130,
	26: 140, 27: 150, 28: 160, 29: 170, 30: 180,
	31: 200, 32: 220, 33: 240, 34: 260, 35: 280,
	36: 300, 37: 320, 38: 340, 39: 360, 40: 380,
	41: 400, 42: 420, 43: 440, 44: 460, 45: 480,
	46: 500, 47: 520, 48: 540, 49: 560, 50: 580,
	51: 600, 52: 620, 53: 640, 54: 660, 55: 680,
	56: 700, 57: 720, 58: 740, 59: 760, 60: 780,
	61: 750, 62: 800, 63: 850, 64: 900, 65: 950,
	66: 1000, 67: 1050, 68: 1100, 69: 1150, 70: 1200,
	71: 1250, 72: 1300, 73: 1350, 74: 1400, 75: 1450,
	76: 1500, 77: 1550, 78: 1600, 79: 1650, 80: 1700,
	81: 1775, 82: 1850, 83: 1925, 84: 2000, 85: 2075,
	86: 2150, 87: 2225, 88: 2300, 89: 2375, 90: 2450,
	91: 2550, 92: 2650, 93: 2750, 94: 2850, 95: 2950,
	96: 3050, 97: 3150, 98: 3250, 99: 3350, 100: 3450,
}

type ChatDomain interface {
	GetMessages(context.Context, *model.GetMessagesRequest) (*model.GetMessagesResponse, error)
	GetChannels(context.Context, *model.GetChannelsRequest) (*model.GetChannelsResponse, error)
	CreateChannel(context.Context, *model.CreateChannelRequest) (*model.CreateChannelResponse, error)
	CreateMessage(context.Context, *model.CreateMessageRequest) (*model.CreateMessageResponse, error)
	DeleteMessage(context.Context, *model.DeleteMessageRequest) (*model.DeleteMessageResponse, error)
	AddReaction(context.Context, *model.AddReactionRequest) (*model.AddReactionResponse, error)
	RemoveReaction(context.Context, *model.RemoveReactionRequest) (*model.RemoveReactionResponse, error)
	GetUserReactions(context.Context, *model.GetUserReactionsRequest) (*model.GetUserReactionsResponse, error)
	DeleteChannel(context.Context, *model.DeleteChannelRequest) (*model.DeleteChannelResponse, error)
	UpdateChannel(context.Context, *model.UpdateChannelRequest) (*model.UpdateChannelResponse, error)
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
	redisClient              xredis.Client
}

func NewChatDomain(
	communityRepo repository.CommunityRepository,
	chatMessageRepo repository.ChatMessageRepository,
	chatChannelRepo repository.ChatChannelRepository,
	chatReactionRepo repository.ChatReactionRepository,
	chatMemberRepo repository.ChatMemberRepository,
	chatChannelBucketRepo repository.ChatChannelBucketRepository,
	userRepo repository.UserRepository,
	followerRepo repository.FollowerRepository,
	notificationEngineCaller client.NotificationEngineCaller,
	roleVerifier *common.CommunityRoleVerifier,
	redisClient xredis.Client,
) *chatDomain {
	return &chatDomain{
		communityRepo:            communityRepo,
		chatMessageRepo:          chatMessageRepo,
		chatChannelRepo:          chatChannelRepo,
		chatReactionRepo:         chatReactionRepo,
		chatMemberRepo:           chatMemberRepo,
		chatChannelBucketRepo:    chatChannelBucketRepo,
		userRepo:                 userRepo,
		followerRepo:             followerRepo,
		roleVerifier:             roleVerifier,
		notificationEngineCaller: notificationEngineCaller,
		redisClient:              redisClient,
	}
}

func (d *chatDomain) GetChannels(
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
		clientChannels = append(clientChannels, model.ConvertChatChannel(&c, req.CommunityHandle))
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
		Description:   req.Description,
		LastMessageID: 0,
	}

	if err := d.chatChannelRepo.Create(ctx, channel); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create channel: %v", err)
		return nil, errorx.Unknown
	}

	go func() {
		ev := event.New(
			&event.ChannelCreatedEvent{ChatChannel: model.ConvertChatChannel(channel, community.Handle)},
			&event.Metadata{ToCommunities: []string{channel.CommunityID}},
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

	userID := xcontext.RequestUserID(ctx)
	user, err := d.userRepo.GetByID(ctx, userID)
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
		AuthorID:    userID,
		ChannelID:   req.ChannelID,
		Content:     req.Content,
		Attachments: req.Attachments,
	}

	if err := d.chatMessageRepo.Create(ctx, &msg); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create message: %v", err)
		return nil, errorx.Unknown
	}

	chatConfig := xcontext.Configs(ctx).Chat
	xp := chatConfig.MessageXP
	for _, attach := range req.Attachments {
		if strings.HasPrefix(attach.ContentType, "image/") && xp < chatConfig.ImageMessageXP {
			xp = chatConfig.ImageMessageXP
		} else if strings.HasPrefix(attach.ContentType, "video/") && xp < chatConfig.VideoMessageXP {
			xp = chatConfig.VideoMessageXP
		}
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
		UserID:            userID,
		ChannelID:         req.ChannelID,
		LastReadMessageID: msg.ID,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update last read message of member: %v", err)
		return nil, errorx.Unknown
	}

	go func() {
		if err = d.increaseChatXP(ctx, userID, channel.CommunityID, xp); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot increase chat xp: %v", err)
		}

		userInfo := model.ConvertShortUser(user, "")
		b, err := d.redisClient.Exist(ctx, common.RedisKeyUserStatus(userID))
		if err != nil {
			xcontext.Logger(ctx).Warnf("Cannot check user status key: %v", err)
		} else {
			if b {
				userInfo.Status = string(event.Online)
			} else {
				userInfo.Status = string(event.Offline)
			}
		}

		ev := event.New(
			&event.MessageCreatedEvent{
				ChatMessage: model.ConvertChatMessage(&msg, userInfo, nil),
			},
			&event.Metadata{ToCommunities: []string{channel.CommunityID}},
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
		err = d.increaseChatXP(ctx, userID, channel.CommunityID, xcontext.Configs(ctx).Chat.ReactionXP)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot increase chat xp: %v", err)
		}

		ev := event.New(
			&event.ReactionAddedEvent{
				ChannelID: req.ChannelID,
				MessageID: req.MessageID,
				UserID:    xcontext.RequestUserID(ctx),
				Emoji:     req.Emoji,
			},
			&event.Metadata{ToCommunities: []string{channel.CommunityID}},
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
			&event.Metadata{ToCommunities: []string{channel.CommunityID}},
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
			xcontext.Logger(ctx).Warnf("Not found author info %s", msg.AuthorID)
			continue
		}

		msgResp = append(msgResp, model.ConvertChatMessage(
			&msg, model.ConvertShortUser(&author, ""), reactionStates[msg.ID]))
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
		return &model.GetUserReactionsResponse{Users: []model.ShortUser{}}, nil
	}

	users, err := d.userRepo.GetByIDs(ctx, reaction.UserIds[:req.Limit])
	if err != nil {
		xcontext.Logger(ctx).Errorf("Unable to get users: %v", err)
		return nil, errorx.Unknown
	}

	respUsers := make([]model.ShortUser, 0, len(reaction.UserIds))
	for _, u := range users {
		respUsers = append(respUsers, model.ConvertShortUser(&u, ""))
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
			&event.Metadata{ToCommunities: []string{channel.CommunityID}},
		)
		if err := d.notificationEngineCaller.Emit(ctx, ev); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot emit delete message event: %v", err)
		}
	}()

	return &model.DeleteMessageResponse{}, nil
}

func (d *chatDomain) DeleteChannel(ctx context.Context, req *model.DeleteChannelRequest) (*model.DeleteChannelResponse, error) {
	channel, err := d.chatChannelRepo.GetByID(ctx, req.ChannelID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Unable to get channel: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.roleVerifier.Verify(ctx, channel.CommunityID); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	if err := d.chatChannelRepo.DeleteByID(ctx, req.ChannelID); err != nil {
		xcontext.Logger(ctx).Errorf("Unable to delete channel: %v", err)
		return nil, errorx.Unknown
	}

	return &model.DeleteChannelResponse{}, nil
}

func (d *chatDomain) increaseChatXP(ctx context.Context, userID, communityID string, xp int) error {
	err := d.followerRepo.IncreaseChatXP(ctx, userID, communityID, xp)
	if err != nil {
		return err
	}

	follower, err := d.followerRepo.Get(ctx, userID, communityID)
	if err != nil {
		return err
	}

	for {
		if follower.ChatLevel >= len(chatLevelConfigs) {
			break
		}

		level := follower.ChatLevel + 1
		thresholdXP, ok := chatLevelConfigs[level]
		if !ok {
			return err
		}

		if follower.CurrentChatXP < thresholdXP {
			break
		}

		err = d.followerRepo.UpdateChatLevel(ctx, userID, communityID, level, thresholdXP)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// In this case, the XP is not enough to update level.
				break
			}

			return err
		}

		follower.ChatLevel += 1
		follower.CurrentChatXP -= thresholdXP
	}

	return nil
}

func (d *chatDomain) UpdateChannel(ctx context.Context, req *model.UpdateChannelRequest) (*model.UpdateChannelResponse, error) {
	channel, err := d.chatChannelRepo.GetByID(ctx, req.ChannelID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Unable to get channel: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.roleVerifier.Verify(ctx, channel.CommunityID); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	if err := d.chatChannelRepo.Update(ctx, &entity.ChatChannel{
		SnowFlakeBase: entity.SnowFlakeBase{
			ID: req.ChannelID,
		},
		Name:        req.ChannelName,
		Description: req.Description,
	}); err != nil {
		xcontext.Logger(ctx).Errorf("Unable to delete channel: %v", err)
		return nil, errorx.Unknown
	}

	return &model.UpdateChannelResponse{}, nil
}
