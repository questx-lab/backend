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
	GetFirstMap(ctx context.Context) (*entity.GameMap, error)
	GetMaps(context.Context) ([]entity.GameMap, error)
	GetMapByID(context.Context, string) (*entity.GameMap, error)
	GetMapByName(context.Context, string) (*entity.GameMap, error)
	GetMapByIDs(context.Context, []string) ([]entity.GameMap, error)
	DeleteMap(context.Context, string) error
	CreateGameTileset(context.Context, *entity.GameMapTileset) error
	GetTilesetsByMapID(context.Context, string) ([]entity.GameMapTileset, error)
	CreateGamePlayer(context.Context, *entity.GameMapPlayer) error
	GetPlayersByMapID(context.Context, string) ([]entity.GameMapPlayer, error)
	GetPlayer(ctx context.Context, name string, mapID string) (*entity.GameMapPlayer, error)
	CreateRoom(context.Context, *entity.GameRoom) error
	GetAllRooms(ctx context.Context) ([]entity.GameRoom, error)
	GetRoomByID(context.Context, string) (*entity.GameRoom, error)
	GetRoomsByCommunityID(context.Context, string) ([]entity.GameRoom, error)
	DeleteRoom(context.Context, string) error
	UpdateRoomEngine(ctx context.Context, roomID, engineID string) error
	CountActiveUsersByRoomID(context.Context, string) (int64, error)
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

func (r *gameRepository) CreateGameTileset(ctx context.Context, data *entity.GameMapTileset) error {
	return xcontext.DB(ctx).Create(data).Error
}

func (r *gameRepository) CreateGamePlayer(ctx context.Context, data *entity.GameMapPlayer) error {
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

func (r *gameRepository) GetTilesetsByMapID(ctx context.Context, mapID string) ([]entity.GameMapTileset, error) {
	result := []entity.GameMapTileset{}
	err := xcontext.DB(ctx).
		Model(&entity.GameMapTileset{}).
		Joins("join game_maps on game_maps.id = game_map_tilesets.game_map_id").
		Find(&result, "game_maps.id=?", mapID).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *gameRepository) GetPlayer(ctx context.Context, name string, mapID string) (*entity.GameMapPlayer, error) {
	result := entity.GameMapPlayer{}
	err := xcontext.DB(ctx).
		Model(&entity.GameMapPlayer{}).
		Joins("join game_maps on game_maps.id = game_map_players.game_map_id").
		Take(&result, "game_maps.id=? AND game_map_players.name=?", mapID, name).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *gameRepository) GetPlayersByMapID(ctx context.Context, mapID string) ([]entity.GameMapPlayer, error) {
	result := []entity.GameMapPlayer{}
	err := xcontext.DB(ctx).
		Model(&entity.GameMapPlayer{}).
		Joins("join game_maps on game_maps.id = game_map_players.game_map_id").
		Find(&result, "game_maps.id=?", mapID).Error
	if err != nil {
		return nil, err
	}

	return result, nil
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

func (r *gameRepository) CountActiveUsersByRoomID(ctx context.Context, roomID string) (int64, error) {
	var result int64
	err := xcontext.DB(ctx).Model(&entity.GameUser{}).
		Joins("join game_rooms on game_rooms.id=game_users.room_id").
		Where("game_users.room_id=? AND game_users.is_active=?", roomID, true).
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

func (r *gameRepository) UpdateRoomEngine(ctx context.Context, roomID, engineID string) error {
	return xcontext.DB(ctx).Model(&entity.GameRoom{}).
		Select("started_by").
		Where("id=?", roomID).
		Updates(map[string]any{
			"started_by": engineID,
		}).Error
}
