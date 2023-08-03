package proxy

import (
	"context"
	"encoding/json"

	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/notification/event"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/questx-lab/backend/pkg/xredis"
	"github.com/redis/go-redis/v9"
	"golang.org/x/exp/slices"
)

const maxSeqCheck = 10

type ProxyServer struct {
	router          *Router
	chatMemberRepo  repository.ChatMemberRepository
	chatChannelRepo repository.ChatChannelRepository
	followerRepo    repository.FollowerRepository
	communityRepo   repository.CommunityRepository
	userRepo        repository.UserRepository

	redisClient xredis.Client
}

func NewProxyServer(
	ctx context.Context,
	chatMemberRepo repository.ChatMemberRepository,
	chatChannelRepo repository.ChatChannelRepository,
	followerRepo repository.FollowerRepository,
	communityRepo repository.CommunityRepository,
	userRepo repository.UserRepository,
	redisClient xredis.Client,
	engineCaller client.NotificationEngineCaller,
) *ProxyServer {
	router := NewRouter(ctx, followerRepo, userRepo, engineCaller, redisClient)
	go router.run(ctx)

	return &ProxyServer{
		router:          router,
		chatMemberRepo:  chatMemberRepo,
		chatChannelRepo: chatChannelRepo,
		followerRepo:    followerRepo,
		communityRepo:   communityRepo,
		userRepo:        userRepo,
		redisClient:     redisClient,
	}
}

func (server *ProxyServer) ServeProxy(ctx context.Context, req *model.ServeNotificationProxyRequest) error {
	userID := xcontext.RequestUserID(ctx)

	session := NewUserSession(userID)
	defer session.Leave()

	userHub, err := server.router.GetUserHub(ctx, userID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get hub: %v", err)
		return errorx.Unknown
	}
	session.JoinUser(userHub)

	followers, err := server.followerRepo.GetListByUserID(ctx, userID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot read followers: %v", err)
		return errorx.Unknown
	}

	readyEvent, err := server.generateReadyEvent(ctx, userID, followers)
	if err != nil {
		return err
	}

	session.C <- event.New(readyEvent, nil)

	for _, follower := range followers {
		communityHub, err := server.router.GetCommunityHub(ctx, follower.CommunityID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get hub: %v", err)
			return errorx.Unknown
		}

		session.JoinCommunity(communityHub)
	}

	wsClient := xcontext.WSClient(ctx)
	var clientSeq uint64
	var lastServerSeqs = make([]uint64, 0, maxSeqCheck)
	for {
		select {
		case ev, ok := <-session.C:
			if !ok {
				return errorx.New(errorx.Unavailable, "Sesssion is closed")
			}

			if slices.Contains(lastServerSeqs, ev.Seq) {
				// This session already received this event before, no need to
				// send to client again.
				continue
			}

			if len(lastServerSeqs) > maxSeqCheck {
				lastServerSeqs = lastServerSeqs[1:]
			}
			lastServerSeqs = append(lastServerSeqs, ev.Seq)

			if ev.Op == (event.FollowCommunityEvent{}).Op() {
				var data event.FollowCommunityEvent
				if err := json.Unmarshal(ev.Data, &data); err != nil {
					xcontext.Logger(ctx).Errorf("Cannot decode follow community event: %v", err)
					return errorx.Unknown
				}

				communityHub, err := server.router.GetCommunityHub(ctx, data.CommunityID)
				if err != nil {
					xcontext.Logger(ctx).Errorf("Cannot get hub: %v", err)
					return errorx.Unknown
				}

				session.JoinCommunity(communityHub)
			}

			evResp := event.Format(ev, clientSeq)
			clientSeq++

			b, err := json.Marshal(evResp)
			if err != nil {
				xcontext.Logger(ctx).Warnf("Cannot marshal resp: %v", err)
				continue
			}

			if err := wsClient.Write(b, false); err != nil {
				xcontext.Logger(ctx).Warnf("Cannot send resp to client: %v", err)
				return errorx.Unknown
			}

		case _, ok := <-wsClient.R:
			if !ok {
				return errorx.Unknown
			}

			// No need to handle websocket request messages.
		}
	}
}

func (server *ProxyServer) generateReadyEvent(
	ctx context.Context,
	userID string,
	followers []entity.Follower,
) (*event.ReadyEvent, error) {
	myChatMembers, err := server.chatMemberRepo.GetByUserID(ctx, userID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get members: %v", err)
		return nil, errorx.Unknown
	}

	chatMemberMap := map[int64]entity.ChatMember{}
	for _, member := range myChatMembers {
		chatMemberMap[member.ChannelID] = member
	}

	communityMap := map[string]entity.Community{}
	for _, f := range followers {
		communityMap[f.CommunityID] = entity.Community{}
	}

	communities, err := server.communityRepo.GetByIDs(ctx, common.MapKeys(communityMap))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get communities: %v", err)
		return nil, errorx.Unknown
	}

	for _, c := range communities {
		communityMap[c.ID] = c
	}

	channels, err := server.chatChannelRepo.GetByCommunityIDs(ctx, common.MapKeys(communityMap))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get channels: %v", err)
		return nil, errorx.Unknown
	}

	chatMembers := []model.ChatMember{}
	for _, channel := range channels {
		member := entity.ChatMember{UserID: userID, ChannelID: channel.ID}
		if m, ok := chatMemberMap[channel.ID]; ok {
			member = m
		}

		community, ok := communityMap[channel.CommunityID]
		if !ok {
			xcontext.Logger(ctx).Errorf("Not found community %s", channel.CommunityID)
			return nil, errorx.Unknown
		}

		chatMembers = append(chatMembers,
			model.ConvertChatMember(&member, model.ConvertChatChannel(&channel, community.Handle)))
	}

	clientCommunities := []model.Community{}
	for _, c := range communities {
		result := model.ConvertCommunity(&c, 0)

		onlineUserIDs, err := server.redisClient.SMembers(ctx, common.RedisKeyCommunityOnline(c.ID), 500)
		if err != nil && err != redis.Nil {
			xcontext.Logger(ctx).Errorf("Cannot get online member of community %s: %v", c.ID, err)
			return nil, errorx.Unknown
		}

		onlineUsers, err := server.userRepo.GetByIDs(ctx, onlineUserIDs)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get online users info: %v", err)
			return nil, errorx.Unknown
		}

		for _, u := range onlineUsers {
			clientUser := model.ConvertShortUser(&u, "")
			b, err := server.redisClient.Exist(ctx, common.RedisKeyUserStatus(u.ID))
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot get user status from redis: %v", err)
				return nil, errorx.Unknown
			}

			if b {
				clientUser.Status = string(event.Online)
			} else {
				clientUser.Status = string(event.Offline)
			}

			result.ChatMembers = append(result.ChatMembers, clientUser)
		}

		clientCommunities = append(clientCommunities, result)
	}

	return &event.ReadyEvent{
		ChatMembers: chatMembers,
		Communities: clientCommunities,
	}, nil
}
