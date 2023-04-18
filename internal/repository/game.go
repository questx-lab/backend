package repository

import (
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type GameRepository interface {
	CreateMap(xcontext.Context, *entity.GameMap) error
	CreateRoom(xcontext.Context, *entity.GameRoom) error
	GetRoomByID(xcontext.Context, string) (*entity.GameRoom, error)
	GetMapByID(xcontext.Context, string) (*entity.GameMap, error)
	GetUsersByRoomID(xcontext.Context, string) ([]entity.GameUser, error)
	UpdateGameUserByID(xcontext.Context, entity.GameUser) error
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
		Take(&result, "game_users.room_id=?", roomID).Error

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *gameRepository) UpdateGameUserByID(ctx xcontext.Context, user entity.GameUser) error {
	err := ctx.DB().Updates(user).Error
	if err != nil {
		return err
	}

	return nil
}
