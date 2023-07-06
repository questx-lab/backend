package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm/clause"
)

type GameCharacterRepository interface {
	Create(ctx context.Context, character *entity.GameCharacter) error
	Get(ctx context.Context, name string, level int) (*entity.GameCharacter, error)
	GetByID(ctx context.Context, characterID string) (*entity.GameCharacter, error)
	GetByName(ctx context.Context, name string) ([]entity.GameCharacter, error)
	GetAll(ctx context.Context) ([]entity.GameCharacter, error)
	Delete(ctx context.Context, name string, level int) error
	CreateCommunityCharacter(ctx context.Context, character *entity.GameCommunityCharacter) error
	GetCommunityCharacter(ctx context.Context, communityID, characterID string) (*entity.GameCommunityCharacter, error)
	GetAllCommunityCharacters(ctx context.Context, communityID string) ([]entity.GameCommunityCharacter, error)
	GetCommunityCharactersByName(ctx context.Context, communityID, name string) ([]entity.GameCommunityCharacter, error)
	GetCommunityCharactersByLevel(ctx context.Context, communityID string, level int) ([]entity.GameCommunityCharacter, error)
	CreateUserCharacter(ctx context.Context, userCharacter *entity.GameUserCharacter) error
	GetAllUserCharacters(ctx context.Context, userID, communityID string) ([]entity.GameUserCharacter, error)
	GetLastLevelUserCharacter(ctx context.Context, userID, communityID, characterName string) (*entity.GameUserCharacter, error)
}

type gameCharacterRepository struct{}

func NewGameCharacterRepository() *gameCharacterRepository {
	return &gameCharacterRepository{}
}

func (r *gameCharacterRepository) Create(ctx context.Context, character *entity.GameCharacter) error {
	return xcontext.DB(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "name"},
				{Name: "level"},
			},
			DoUpdates: clause.Assignments(map[string]any{
				"config_url":          character.ConfigURL,
				"image_url":           character.ImageURL,
				"thumbnail_url":       character.ThumbnailURL,
				"sprite_width_ratio":  character.SpriteWidthRatio,
				"sprite_height_ratio": character.SpriteHeightRatio,
			}),
		}).Create(character).Error
}

func (r *gameCharacterRepository) Get(ctx context.Context, name string, level int) (*entity.GameCharacter, error) {
	var result *entity.GameCharacter
	if err := xcontext.DB(ctx).Take(&result, "name=? AND level=?", name, level).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *gameCharacterRepository) GetByID(ctx context.Context, characterID string) (*entity.GameCharacter, error) {
	var result *entity.GameCharacter
	if err := xcontext.DB(ctx).Take(&result, "id=?", characterID).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *gameCharacterRepository) GetByName(ctx context.Context, name string) ([]entity.GameCharacter, error) {
	var result []entity.GameCharacter
	if err := xcontext.DB(ctx).Find(&result, "name=?", name).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *gameCharacterRepository) GetAll(ctx context.Context) ([]entity.GameCharacter, error) {
	var result []entity.GameCharacter
	if err := xcontext.DB(ctx).Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *gameCharacterRepository) Delete(ctx context.Context, name string, level int) error {
	return xcontext.DB(ctx).
		Where("name=? AND level=?", name, level).
		Delete(&entity.GameCharacter{}).Error
}

func (r *gameCharacterRepository) CreateCommunityCharacter(
	ctx context.Context, character *entity.GameCommunityCharacter,
) error {
	return xcontext.DB(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "community_id"},
				{Name: "character_id"},
			},
			DoUpdates: clause.Assignments(map[string]any{
				"points": character.Points,
			}),
		}).Create(character).Error
}

func (r *gameCharacterRepository) GetCommunityCharacter(
	ctx context.Context, communityID, characterID string,
) (*entity.GameCommunityCharacter, error) {
	var result entity.GameCommunityCharacter
	err := xcontext.DB(ctx).
		Where("community_id=? AND character_id=?", communityID, characterID).
		Take(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *gameCharacterRepository) GetAllCommunityCharacters(
	ctx context.Context, communityID string,
) ([]entity.GameCommunityCharacter, error) {
	var result []entity.GameCommunityCharacter
	if err := xcontext.DB(ctx).Where("community_id=?", communityID).Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *gameCharacterRepository) GetCommunityCharactersByName(
	ctx context.Context, communityID, name string,
) ([]entity.GameCommunityCharacter, error) {
	var result []entity.GameCommunityCharacter
	if err := xcontext.DB(ctx).
		Joins("join game_characters on game_characters.id=game_community_characters.character_id").
		Find(&result, "game_characters.name=?", name).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *gameCharacterRepository) GetCommunityCharactersByLevel(
	ctx context.Context, communityID string, level int,
) ([]entity.GameCommunityCharacter, error) {
	var result []entity.GameCommunityCharacter
	if err := xcontext.DB(ctx).
		Joins("join game_characters on game_characters.id=game_community_characters.character_id").
		Find(&result, "game_characters.level=?", level).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *gameCharacterRepository) CreateUserCharacter(
	ctx context.Context, userCharacter *entity.GameUserCharacter,
) error {
	return xcontext.DB(ctx).Create(userCharacter).Error
}

func (r *gameCharacterRepository) GetAllUserCharacters(
	ctx context.Context, userID, communityID string,
) ([]entity.GameUserCharacter, error) {
	var result []entity.GameUserCharacter
	err := xcontext.DB(ctx).Model(&entity.GameUserCharacter{}).
		Joins("join game_characters on game_characters.id=game_user_characters.character_id").
		Where("game_user_characters.user_id=?", userID).
		Where("game_user_characters.community_id=?", communityID).
		Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *gameCharacterRepository) GetLastLevelUserCharacter(
	ctx context.Context, userID, communityID, characterName string,
) (*entity.GameUserCharacter, error) {
	var result entity.GameUserCharacter
	err := xcontext.DB(ctx).Model(&entity.GameUserCharacter{}).
		Joins("join game_characters on game_characters.id=game_user_characters.character_id").
		Where("game_user_characters.user_id=?", userID).
		Where("game_user_characters.community_id=?", communityID).
		Where("game_characters.name=?", characterName).
		Order("game_characters.level DESC").
		Find(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}
