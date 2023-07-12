package repository

import (
	"context"
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm/clause"
)

type GameRepository interface {
	UpsertMap(context.Context, *entity.GameMap) error
	GetFirstMap(ctx context.Context) (*entity.GameMap, error)
	GetMaps(context.Context) ([]entity.GameMap, error)
	GetMapByID(context.Context, string) (*entity.GameMap, error)
	GetMapByName(context.Context, string) (*entity.GameMap, error)
	GetMapByIDs(context.Context, []string) ([]entity.GameMap, error)
	DeleteMap(context.Context, string) error
	CreateRoom(context.Context, *entity.GameRoom) error
	GetAllRooms(ctx context.Context) ([]entity.GameRoom, error)
	GetRoomByID(context.Context, string) (*entity.GameRoom, error)
	GetRoomsByCommunityID(context.Context, string) ([]entity.GameRoom, error)
	DeleteRoom(context.Context, string) error
	UpdateRoomEngine(ctx context.Context, roomID, engineID string) error
	GetRoomsByUserCommunity(ctx context.Context, userID, communityID string) ([]entity.GameRoom, error)
	CountActiveUsersByRoomID(ctx context.Context, roomID, excludedUserID string) (int64, error)
	GetUsersByRoomID(context.Context, string) ([]entity.GameUser, error)
	GetUser(ctx context.Context, userID, roomID string) (*entity.GameUser, error)
	UpsertGameUser(context.Context, *entity.GameUser) error
	UpsertGameUserWithProxy(context.Context, *entity.GameUser) error
}

type gameRepository struct{}

func NewGameRepository() *gameRepository {
	return &gameRepository{}
}

func (r *gameRepository) UpsertMap(ctx context.Context, data *entity.GameMap) error {
	return xcontext.DB(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "id"},
			},
			DoUpdates: clause.Assignments(map[string]any{
				"config_url": data.ConfigURL,
				"name":       data.Name,
			}),
		}).Create(data).Error
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

func (r *gameRepository) GetFirstMap(ctx context.Context) (*entity.GameMap, error) {
	result := entity.GameMap{}
	if err := xcontext.DB(ctx).Order("created_at ASC").Take(&result).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *gameRepository) GetMapByName(ctx context.Context, name string) (*entity.GameMap, error) {
	result := entity.GameMap{}
	if err := xcontext.DB(ctx).Take(&result, "name=?", name).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *gameRepository) GetMapByIDs(ctx context.Context, mapIDs []string) ([]entity.GameMap, error) {
	result := []entity.GameMap{}
	if err := xcontext.DB(ctx).Find(&result, "id IN (?)", mapIDs).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *gameRepository) GetUsersByRoomID(ctx context.Context, roomID string) ([]entity.GameUser, error) {
	result := []entity.GameUser{}
	err := xcontext.DB(ctx).Model(&entity.GameUser{}).
		Find(&result, "room_id=?", roomID).Error

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *gameRepository) GetUser(ctx context.Context, userID, roomID string) (*entity.GameUser, error) {
	result := entity.GameUser{}
	err := xcontext.DB(ctx).Model(&entity.GameUser{}).
		Take(&result, "user_id=? AND room_id=?", userID, roomID).Error

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *gameRepository) CountActiveUsersByRoomID(ctx context.Context, roomID, excludedUserID string) (int64, error) {
	var result int64
	err := xcontext.DB(ctx).Model(&entity.GameUser{}).
		Joins("join game_rooms on game_rooms.id=game_users.room_id").
		Where("game_users.room_id=? AND game_users.connected_by != ?", roomID, nil).
		Where("game_users.user_id != ?", excludedUserID).
		Count(&result).Error

	if err != nil {
		return 0, err
	}

	return result, nil
}

func (r *gameRepository) GetMaps(ctx context.Context) ([]entity.GameMap, error) {
	var result []entity.GameMap
	if err := xcontext.DB(ctx).Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *gameRepository) GetRoomsByCommunityID(ctx context.Context, communityID string) ([]entity.GameRoom, error) {
	var result []entity.GameRoom
	if err := xcontext.DB(ctx).Where("community_id=?", communityID).Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *gameRepository) GetAllRooms(ctx context.Context) ([]entity.GameRoom, error) {
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
				"position_x":   user.PositionX,
				"position_y":   user.PositionY,
				"direction":    user.Direction,
				"connected_by": user.ConnectedBy,
			}),
		},
	).Create(user).Error
}

func (r *gameRepository) UpsertGameUserWithProxy(ctx context.Context, user *entity.GameUser) error {
	return xcontext.DB(ctx).Clauses(
		clause.OnConflict{
			Columns: []clause.Column{
				{Name: "user_id"},
				{Name: "room_id"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"connected_by": user.ConnectedBy,
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

func (r *gameRepository) UpdateRoomEngine(ctx context.Context, roomID, engineID string) error {
	return xcontext.DB(ctx).Model(&entity.GameRoom{}).
		Select("started_by").
		Where("id=?", roomID).
		Updates(map[string]any{"started_by": engineID}).Error
}

func (r *gameRepository) GetRoomsByUserCommunity(ctx context.Context, userID, communityID string) ([]entity.GameRoom, error) {
	result := []entity.GameRoom{}
	err := xcontext.DB(ctx).Model(&entity.GameRoom{}).
		Joins("join game_users on game_rooms.id=game_users.room_id").
		Find(&result, "game_users.user_id=? AND game_rooms.community_id=?", userID, communityID).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}
