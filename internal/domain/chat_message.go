package domain

import (
	"context"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"

	"github.com/bwmarrin/snowflake"
)

type ChatMessageDomain interface {
	GetList(ctx context.Context, req *model.GetListMessageRequest) (*model.GetListMessageResponse, error)
}

type chatMessageDomain struct {
	chatMessageRepo                  repository.ChatMessageRepository
	chatMessageReactionStatisticRepo repository.ChatMessageReactionStatisticRepository
	idGenerator                      snowflake.ID
}

func NewChatMessageDomain(
	chatMessageRepo repository.ChatMessageRepository,
	chatMessageReactionStatisticRepo repository.ChatMessageReactionStatisticRepository,
	idGenerator snowflake.ID,
) ChatMessageDomain {
	return &chatMessageDomain{
		chatMessageRepo:                  chatMessageRepo,
		chatMessageReactionStatisticRepo: chatMessageReactionStatisticRepo,
		idGenerator:                      idGenerator,
	}
}

func (d *chatMessageDomain) GetList(ctx context.Context, req *model.GetListMessageRequest) (*model.GetListMessageResponse, error) {

	channelID, err := snowflake.ParseString(req.ChannelID)
	if err != nil {
		return nil, err
	}

	lastMessageID, err := snowflake.ParseString(req.LastMessageID)
	if err != nil {
		return nil, err
	}
	fromBucket := lastMessageID.Time() / common.BucketDuration.Nanoseconds()
	toBucket := channelID.Time() / common.BucketDuration.Nanoseconds()
	messages, err := d.chatMessageRepo.GetListByLastMessage(ctx, &repository.LastMessageFilter{
		ChannelID:     req.ChannelID,
		LastMessageID: req.LastMessageID,
		Limit:         req.Limit,
		FromBucket:    fromBucket,
		ToBucket:      toBucket,
	})

	messageIDs := []string{}
	for _, mess := range messages {
		messageIDs = append(messageIDs, mess.ID)
	}

	reactions, err := d.chatMessageReactionStatisticRepo.GetListByMessages(ctx, messageIDs)
	if err != nil {
		return nil, err
	}

	m := make(map[string][]model.ChatReaction)

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
