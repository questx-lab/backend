package gameproxy

import (
	"context"
	"encoding/json"
	"time"

	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/logger"
	"github.com/questx-lab/backend/pkg/pubsub"
	"github.com/questx-lab/backend/pkg/ws"
)

type ResponseSubscribeHandler interface {
	Subscribe(ctx context.Context, pack *pubsub.Pack, t time.Time)
}

type responseSubscribeHandler struct {
	hub    *ws.Hub
	logger logger.Logger
}

func NewResponseSubscribeHandler(hub *ws.Hub, logger logger.Logger) ResponseSubscribeHandler {
	return &responseSubscribeHandler{
		hub:    hub,
		logger: logger,
	}
}
func (s *responseSubscribeHandler) Subscribe(ctx context.Context, pack *pubsub.Pack, t time.Time) {
	var resp model.GameActionServerResponse
	if err := json.Unmarshal(pack.Msg, &resp); err != nil {
		s.logger.Errorf("Unable to unmarshal: %v", err)
		return
	}
	clientResp := model.GameActionClientResponse{
		Type:  resp.Type,
		Value: resp.Value,
	}
	b, err := json.Marshal(clientResp)
	if err != nil {
		s.logger.Errorf("Unable to marshal action client: %v", err)
		return
	}

	roomID := string(pack.Key)

	s.hub.BroadCastByRoomID(roomID, b)
}
