package entity

import (
	"time"

	"gorm.io/gorm"
)

type GameCharacter struct {
	Base
	Name              string `gorm:"index:idx_game_characters_name_level,unique"`
	Level             int    `gorm:"index:idx_game_characters_name_level,unique"`
	ConfigURL         string
	ImageURL          string
	ThumbnailURL      string
	SpriteWidthRatio  float64
	SpriteHeightRatio float64
	Points            int
}

type GameCommunityCharacter struct {
	CommunityID string        `gorm:"primaryKey"`
	Community   Community     `gorm:"foreignKey:CommunityID"`
	CharacterID string        `gorm:"primaryKey"`
	Character   GameCharacter `gorm:"foreignKey:CharacterID"`
	Points      int

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

type GameUserCharacter struct {
	UserID      string        `gorm:"primaryKey"`
	User        User          `gorm:"foreignKey:UserID"`
	CommunityID string        `gorm:"primaryKey"`
	Community   Community     `gorm:"foreignKey:CommunityID"`
	CharacterID string        `gorm:"primaryKey"`
	Character   GameCharacter `gorm:"foreignKey:CharacterID"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}
