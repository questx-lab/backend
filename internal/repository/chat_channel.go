package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/scylladb/gocqlx/v2"
)

type ChatChannelRepository interface {
	Create(ctx context.Context, data *entity.ChatChannel) error
	GetByID(ctx context.Context, id int64) (*entity.ChatChannel, error)
	UpdateLastMessageByID(ctx context.Context, id, lastMessageID int64) error
	DeleteByID(ctx context.Context, id int64) error
	CountByCommunityID(ctx context.Context, communityID string) (int64, error)
}

type chatChannelRepository struct{}

func NewChatChannelRepository(session gocqlx.Session) ChatChannelRepository {
	return &chatChannelRepository{}
}

func (r *chatChannelRepository) Create(ctx context.Context, data *entity.ChatChannel) error {
	if err := xcontext.DB(ctx).Create(data).Error; err != nil {
		return err
	}

	return nil
}

func (r *chatChannelRepository) GetByID(ctx context.Context, id int64) (*entity.ChatChannel, error) {
	var result entity.ChatChannel
	if err := xcontext.DB(ctx).Take(&result, "id=?", id).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *chatChannelRepository) UpdateLastMessageByID(ctx context.Context, id, lastMessageID int64) error {
	if err := xcontext.DB(ctx).Model(&entity.ChatChannel{}).
		Where("id=? AND last_message_id<?", id, lastMessageID).
		Update("last_message_id", lastMessageID).Error; err != nil {
		return err
	}

	return nil
}

func (r *chatChannelRepository) DeleteByID(ctx context.Context, id int64) error {
	if err := xcontext.DB(ctx).Delete(&entity.ChatChannel{}, "id=?", id).Error; err != nil {
		return err
	}

	return nil
}

func (r *chatChannelRepository) CountByCommunityID(ctx context.Context, communityID string) (int64, error) {
	var result int64
	err := xcontext.DB(ctx).Model(&entity.ChatChannel{}).
		Where("community_id=?", communityID).
		Count(&result).Error
	if err != nil {
		return 0, err
	}

	return result, nil
}
