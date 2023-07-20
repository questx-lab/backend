package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/questx-lab/backend/internal/domain/notification/directive"
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
}

func NewProxyServer(
	ctx context.Context,
	chatMemberRepo repository.ChatMemberRepository,
	chatChannelRepo repository.ChatChannelRepository,
	followerRepo repository.FollowerRepository,
) *ProxyServer {
	return &ProxyServer{
		router:          NewRouter(ctx),
		chatMemberRepo:  chatMemberRepo,
		chatChannelRepo: chatChannelRepo,
		followerRepo:    followerRepo,
	}
}

func (server *ProxyServer) ServeProxy(ctx context.Context, req *model.ServeNotificationProxyRequest) error {
	session := NewSession()
	defer session.LeaveAllHubs()

	userID := xcontext.RequestUserID(ctx)
	myChatMembers, err := server.chatMemberRepo.GetByUserID(ctx, xcontext.RequestUserID(ctx))
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

	comunityIDs := []string{}
	for _, f := range followers {
		comunityIDs = append(comunityIDs, f.CommunityID)
	}

	channels, err := server.chatChannelRepo.GetByCommunityIDs(ctx, comunityIDs)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot read channels: %v", err)
		return errorx.Unknown
	}

	chatMember := []model.ChatMember{}
	for _, channel := range channels {
		member := entity.ChatMember{UserID: userID, ChannelID: channel.ID}
		if m, ok := chatMemberMap[channel.ID]; ok {
			member = m
		}

		chatMember = append(chatMember, model.ChatMember{
			UserID: userID,
			Channel: model.ChatChannel{
				ID:            channel.ID,
				CommunityID:   channel.CommunityID,
				Name:          channel.Name,
				LastMessageID: channel.LastMessageID,
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
		hub, err := server.router.GetHub(ctx, follower.CommunityID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get hub: %v", err)
			return errorx.Unknown
		}

		session.JoinHub(hub)
	}

	wsClient := xcontext.WSClient(ctx)
	var seq int64
	for {
		select {
		case ev, ok := <-session.C:
			if !ok {
				return errorx.New(errorx.Unavailable, "Sesssion is closed")
			}

			x := rand.Intn(2000)

			start := time.Now()
			evResp := event.Format(ev, seq)
			seq++

			fmt.Println("XXXX", evResp.Data)
			b, err := json.Marshal(evResp)
			if err != nil {
				xcontext.Logger(ctx).Warnf("Cannot marshal resp: %v", err)
				continue
			}
			fmt.Println("Y", string(b))

			if err := wsClient.Write(b, false); err != nil {
				xcontext.Logger(ctx).Warnf("Cannot send resp to client: %v", err)
				return errorx.Unknown
			}
			if x < 5 {
				fmt.Println(time.Since(start))
			}

		case req, ok := <-wsClient.R:
			if !ok {
				return errorx.Unknown
			}

			var d directive.ServerDirective
			if err := json.Unmarshal(req, &d); err != nil {
				xcontext.Logger(ctx).Errorf("Cannot unmarshal directive: %v", err)
				return errorx.New(errorx.BadRequest, "Invalid directive")
			}

			switch d.Op {
			case directive.ProxyPingDirectiveOp:
			}
		}
	}
}
