package domain

import (
	"context"

	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/idutil"
)

type ChatDomain interface {
	GetList(ctx context.Context, req *model.GetListMessageRequest) (*model.GetListMessageResponse, error)
}

type chatDomain struct {
	chatMessageRepo                  repository.ChatMessageRepository
	chatMessageReactionStatisticRepo repository.ChatMessageReactionStatisticRepository
}

func NewChatDomain(
	chatMessageRepo repository.ChatMessageRepository,
	chatMessageReactionStatisticRepo repository.ChatMessageReactionStatisticRepository,
) ChatDomain {
	return &chatDomain{
		chatMessageRepo:                  chatMessageRepo,
		chatMessageReactionStatisticRepo: chatMessageReactionStatisticRepo,
	}
}

func (d *chatDomain) GetList(ctx context.Context, req *model.GetListMessageRequest) (*model.GetListMessageResponse, error) {
	if req.LastMessageID == 0 {
		req.LastMessageID = d.snowflakeNode.Generate().Int64()
	}

	fromBucket := idutil.GetBucketByID(req.LastMessageID)
	toBucket := idutil.GetBucketByID(req.ChannelID)
	messages, err := d.chatMessageRepo.GetListByLastMessage(ctx, &repository.LastMessageFilter{
		ChannelID:     req.ChannelID,
		LastMessageID: req.LastMessageID,
		Limit:         req.Limit,
		FromBucket:    fromBucket,
		ToBucket:      toBucket,
	})

	if err != nil {
		return nil, err
	}

	messageIDs := []int64{}
	for _, mess := range messages {
		messageIDs = append(messageIDs, mess.ID)
	}

	reactions, err := d.chatMessageReactionStatisticRepo.GetListByMessages(ctx, messageIDs)
	if err != nil {
		return nil, err
	}

	m := make(map[int64][]model.ChatReaction)

	for _, reaction := range reactions {
		m[reaction.MessageID] = append(m[reaction.MessageID], model.ChatReaction{
			ReactionID: reaction.ReactionID,
			Quantity:   int64(reaction.Quantity),
		})
	}

	var messResp []model.ChatMessage
	for _, mess := range messages {
		cm := model.ChatMessage{
			MessageID: mess.ID,
			Message:   mess.Message,
		}

		if reactions, ok := m[mess.ID]; ok {
			cm.Reactions = reactions
		}
		messResp = append(messResp, cm)
	}

	return &model.GetListMessageResponse{
		Messages: messResp,
	}, nil
}
