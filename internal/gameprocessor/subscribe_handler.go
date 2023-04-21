package gameprocessor

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/logger"
	"github.com/questx-lab/backend/pkg/pubsub"
)

type RequestSubscribeHandler interface {
	Subscribe(ctx context.Context, pack *pubsub.Pack, t time.Time)
}

type requestSubscribeHandler struct {
	publisher pubsub.Publisher
	logger    logger.Logger
}

func NewRequestSubscribeHandler(publisher pubsub.Publisher, logger logger.Logger) RequestSubscribeHandler {
	return &requestSubscribeHandler{
		publisher: publisher,
		logger:    logger,
	}
}
func (s *requestSubscribeHandler) Subscribe(ctx context.Context, pack *pubsub.Pack, t time.Time) {
	log.Printf("%+v", *pack)
	var req model.GameActionServerRequest
	if err := json.Unmarshal(pack.Msg, &req); err != nil {
		s.logger.Errorf("Unable to unmarshal: %v", err)
		return
	}

	//////////////////////// DO IMPLEMENT GAME ACTION HERE ////////////////////////

	resp := model.GameActionServerResponse{
		UserID: req.UserID,
		Type:   req.Type,
		Value:  req.Value,
	}

	b, err := json.Marshal(resp)
	if err != nil {
		s.logger.Errorf("Unable to marshal: %v", err)
		return
	}

	if err := s.publisher.Publish(ctx, model.ResponseTopic, &pubsub.Pack{
		Key: pack.Key, // roomID
		Msg: b,
	}); err != nil {
		s.logger.Errorf("Unable publish by request: %v", err)
		return
	}
}
