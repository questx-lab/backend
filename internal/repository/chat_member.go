package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ChatMemberRepository interface {
	Upsert(ctx context.Context, data *entity.ChatMember) error
	GetByUserID(ctx context.Context, userID string) ([]entity.ChatMember, error)
}

type chatMemberRepository struct{}

func NewChatMemberRepository() ChatMemberRepository {
	return &chatMemberRepository{}
}

func (r *chatMemberRepository) Upsert(ctx context.Context, data *entity.ChatMember) error {
	if err := xcontext.DB(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "user_id"},
				{Name: "channel_id"},
			},
			DoUpdates: clause.Assignments(map[string]any{
				"last_read_message_id": data.LastReadMessageID,
			}),
			Where: clause.Where{
				Exprs: []clause.Expression{gorm.Expr("last_read_message_id<?", data.LastReadMessageID)},
			},
		}).Create(data).Error; err != nil {
		return err
	}

	return nil
}

func (r *chatMemberRepository) GetByUserID(ctx context.Context, userID string) ([]entity.ChatMember, error) {
	var result []entity.ChatMember
	if err := xcontext.DB(ctx).Find(&result, "user_id=?", userID).Error; err != nil {
		return nil, err
	}

	return result, nil
}
