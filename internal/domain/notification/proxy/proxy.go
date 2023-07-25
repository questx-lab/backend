package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/notification/event"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type ProxyServer struct {
	router          *Router
	chatMemberRepo  repository.ChatMemberRepository
	chatChannelRepo repository.ChatChannelRepository
	followerRepo    repository.FollowerRepository
	communityRepo   repository.CommunityRepository
}

func NewProxyServer(
	ctx context.Context,
	chatMemberRepo repository.ChatMemberRepository,
	chatChannelRepo repository.ChatChannelRepository,
	followerRepo repository.FollowerRepository,
	communityRepo repository.CommunityRepository,
) *ProxyServer {
	return &ProxyServer{
		router:          NewRouter(ctx),
		chatMemberRepo:  chatMemberRepo,
		chatChannelRepo: chatChannelRepo,
		followerRepo:    followerRepo,
		communityRepo:   communityRepo,
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

	myChatMembers, err := server.chatMemberRepo.GetByUserID(ctx, userID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get members: %v", err)
		return errorx.Unknown
	}

	chatMemberMap := map[int64]entity.ChatMember{}
	for _, member := range myChatMembers {
		chatMemberMap[member.ChannelID] = member
	}

	followers, err := server.followerRepo.GetListByUserID(ctx, userID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot read followers: %v", err)
		return errorx.Unknown
	}

	communityMap := map[string]entity.Community{}
	for _, f := range followers {
		communityMap[f.CommunityID] = entity.Community{}
	}

	communities, err := server.communityRepo.GetByIDs(ctx, common.MapKeys(communityMap))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get communities: %v", err)
		return errorx.Unknown
	}

	for _, c := range communities {
		communityMap[c.ID] = c
	}

	channels, err := server.chatChannelRepo.GetByCommunityIDs(ctx, common.MapKeys(communityMap))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get channels: %v", err)
		return errorx.Unknown
	}

	chatMember := []model.ChatMember{}
	for _, channel := range channels {
		member := entity.ChatMember{UserID: userID, ChannelID: channel.ID}
		if m, ok := chatMemberMap[channel.ID]; ok {
			member = m
		}

		community, ok := communityMap[channel.CommunityID]
		if !ok {
			xcontext.Logger(ctx).Errorf("Not found community %s", channel.CommunityID)
			return errorx.Unknown
		}

		chatMember = append(chatMember, model.ChatMember{
			UserID: userID,
			Channel: model.ChatChannel{
				ID:              channel.ID,
				CommunityHandle: community.Handle,
				Name:            channel.Name,
				LastMessageID:   channel.LastMessageID,
			},
			LastReadMessageID: member.LastReadMessageID,
		})
	}

	session.C <- event.New(
		&event.ReadyEvent{
			ChatMembers: chatMember,
		},
		&event.Metadata{},
	)

	for _, follower := range followers {
		communityHub, err := server.router.GetCommunityHub(ctx, follower.CommunityID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get hub: %v", err)
			return errorx.Unknown
		}

		session.JoinCommunity(communityHub)
	}

	wsClient := xcontext.WSClient(ctx)
	var seq int64
	for {
		select {
		case ev, ok := <-session.C:
			if !ok {
				return errorx.New(errorx.Unavailable, "Sesssion is closed")
			}

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

			x := rand.Intn(2000)

			start := time.Now()
			evResp := event.Format(ev, seq)
			seq++

			b, err := json.Marshal(evResp)
			if err != nil {
				xcontext.Logger(ctx).Warnf("Cannot marshal resp: %v", err)
				continue
			}

			if err := wsClient.Write(b, false); err != nil {
				xcontext.Logger(ctx).Warnf("Cannot send resp to client: %v", err)
				return errorx.Unknown
			}
			if x < 5 {
				fmt.Println(time.Since(start))
			}

		case _, ok := <-wsClient.R:
			if !ok {
				return errorx.Unknown
			}

			// No need to handle websocket request messages.
		}
	}
}
