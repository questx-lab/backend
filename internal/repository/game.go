package repository

import (
	"context"
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm/clause"
)

type GameRepository interface {
	CreateMap(context.Context, *entity.GameMap) error
	DeleteMap(context.Context, string) error
	CreateRoom(context.Context, *entity.GameRoom) error
	GetRoomByID(context.Context, string) (*entity.GameRoom, error)
	GetMapByID(context.Context, string) (*entity.GameMap, error)
	GetRooms(context.Context) ([]entity.GameRoom, error)
	DeleteRoom(context.Context, string) error
	GetUsersByRoomID(context.Context, string) ([]entity.GameUser, error)
	UpsertGameUser(context.Context, *entity.GameUser) error
}

type gameRepository struct{}

func NewGameRepository() *gameRepository {
	return &gameRepository{}
}

func (r *gameRepository) CreateMap(ctx context.Context, data *entity.GameMap) error {
	return xcontext.DB(ctx).Create(data).Error
}

func (r *gameRepository) CreateRoom(ctx context.Context, data *entity.GameRoom) error {
	return xcontext.DB(ctx).Create(data).Error
}

func (r *gameRepository) GetRoomByID(ctx context.Context, roomID string) (*entity.GameRoom, error) {
	result := entity.GameRoom{}
	if err := xcontext.DB(ctx).Take(&result, "id=?", roomID).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *gameRepository) GetMapByID(ctx context.Context, mapID string) (*entity.GameMap, error) {
	result := entity.GameMap{}
	if err := xcontext.DB(ctx).Take(&result, "id=?", mapID).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *gameRepository) GetUsersByRoomID(ctx context.Context, roomID string) ([]entity.GameUser, error) {
	result := []entity.GameUser{}
	err := xcontext.DB(ctx).Model(&entity.GameUser{}).
		Joins("join game_rooms on game_rooms.id=game_users.room_id").
		Find(&result, "game_users.room_id=?", roomID).Error

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *gameRepository) GetRooms(ctx context.Context) ([]entity.GameRoom, error) {
	var result []entity.GameRoom
	if err := xcontext.DB(ctx).Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *gameRepository) UpsertGameUser(ctx context.Context, user *entity.GameUser) error {
	return xcontext.DB(ctx).Clauses(
		clause.OnConflict{
			Columns: []clause.Column{
				{Name: "user_id"},
				{Name: "room_id"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"position_x": user.PositionX,
				"position_y": user.PositionY,
				"direction":  user.Direction,
				"is_active":  user.IsActive,
			}),
		},
	).Create(user).Error
}

func (r *gameRepository) DeleteMap(ctx context.Context, mapID string) error {
	tx := xcontext.DB(ctx).Delete(&entity.GameMap{}, "id=?", mapID)
	if err := tx.Error; err != nil {
		return err
	}

	if tx.RowsAffected == 0 {
		return fmt.Errorf("row effected is wrong")
	}

	return nil
}

func (r *gameRepository) DeleteRoom(ctx context.Context, roomID string) error {
	tx := xcontext.DB(ctx).Delete(&entity.GameRoom{}, "id=?", roomID)
	if err := tx.Error; err != nil {
		return err
	}

	if tx.RowsAffected == 0 {
		return fmt.Errorf("row effected is wrong")
	}

	return nil
}
