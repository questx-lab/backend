package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type ChatChannelRepository interface {
	Create(ctx context.Context, data *entity.ChatChannel) error
	GetByID(ctx context.Context, id int64) (*entity.ChatChannel, error)
	GetByCommunityID(ctx context.Context, communityID string) ([]entity.ChatChannel, error)
	GetByCommunityIDs(ctx context.Context, communityIDs []string) ([]entity.ChatChannel, error)
	UpdateLastMessageByID(ctx context.Context, id, lastMessageID int64) error
	DeleteByID(ctx context.Context, id int64) error
	CountByCommunityID(ctx context.Context, communityID string) (int64, error)
	Update(ctx context.Context, data *entity.ChatChannel) error
}

type chatChannelRepository struct{}

func NewChatChannelRepository() ChatChannelRepository {
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

func (r *chatChannelRepository) GetByCommunityID(ctx context.Context, communityID string) ([]entity.ChatChannel, error) {
	var result []entity.ChatChannel
	if err := xcontext.DB(ctx).Find(&result, "community_id=?", communityID).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *chatChannelRepository) GetByCommunityIDs(ctx context.Context, communityIDs []string) ([]entity.ChatChannel, error) {
	var result []entity.ChatChannel
	if err := xcontext.DB(ctx).Find(&result, "community_id IN (?)", communityIDs).Error; err != nil {
		return nil, err
	}

	return result, nil
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

func (r *chatChannelRepository) Update(ctx context.Context, data *entity.ChatChannel) error {
	if err := xcontext.DB(ctx).Model(&entity.ChatChannel{}).
		Where("id = ?", data.ID).
		Updates(data).Error; err != nil {
		return err
	}

	return nil
}
