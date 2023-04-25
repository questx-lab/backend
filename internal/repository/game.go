package repository

import (
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm/clause"
)

type GameRepository interface {
	CreateMap(xcontext.Context, *entity.GameMap) error
	DeleteMap(xcontext.Context, string) error
	CreateRoom(xcontext.Context, *entity.GameRoom) error
	GetRoomByID(xcontext.Context, string) (*entity.GameRoom, error)
	GetMapByID(xcontext.Context, string) (*entity.GameMap, error)
	GetRooms(xcontext.Context) ([]entity.GameRoom, error)
	DeleteRoom(xcontext.Context, string) error
	GetUsersByRoomID(xcontext.Context, string) ([]entity.GameUser, error)
	UpsertGameUser(xcontext.Context, *entity.GameUser) error
}

type gameRepository struct{}

func NewGameRepository() *gameRepository {
	return &gameRepository{}
}

func (r *gameRepository) CreateMap(ctx xcontext.Context, data *entity.GameMap) error {
	return ctx.DB().Create(data).Error
}

func (r *gameRepository) CreateRoom(ctx xcontext.Context, data *entity.GameRoom) error {
	return ctx.DB().Create(data).Error
}

func (r *gameRepository) GetRoomByID(ctx xcontext.Context, roomID string) (*entity.GameRoom, error) {
	result := entity.GameRoom{}
	if err := ctx.DB().Take(&result, "id=?", roomID).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *gameRepository) GetMapByID(ctx xcontext.Context, mapID string) (*entity.GameMap, error) {
	result := entity.GameMap{}
	if err := ctx.DB().Take(&result, "id=?", mapID).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *gameRepository) GetUsersByRoomID(ctx xcontext.Context, roomID string) ([]entity.GameUser, error) {
	result := []entity.GameUser{}
	err := ctx.DB().Model(&entity.GameUser{}).
		Joins("join game_rooms on game_rooms.id=game_users.room_id").
		Find(&result, "game_users.room_id=?", roomID).Error

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *gameRepository) GetRooms(ctx xcontext.Context) ([]entity.GameRoom, error) {
	var result []entity.GameRoom
	if err := ctx.DB().Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *gameRepository) UpsertGameUser(ctx xcontext.Context, user *entity.GameUser) error {
	return ctx.DB().Clauses(
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

func (r *gameRepository) DeleteMap(ctx xcontext.Context, mapID string) error {
	tx := ctx.DB().Model(&entity.GameMap{}).Delete("id = ?", mapID)
	if err := tx.Error; err != nil {
		return err
	}

	if tx.RowsAffected == 0 {
		return fmt.Errorf("row effected is wrong")
	}

	return nil
}

func (r *gameRepository) DeleteRoom(ctx xcontext.Context, roomID string) error {
	tx := ctx.DB().Model(&entity.GameRoom{}).Delete("id = ?", roomID)
	if err := tx.Error; err != nil {
		return err
	}

	if tx.RowsAffected == 0 {
		return fmt.Errorf("row effected is wrong")
	}

	return nil
}
