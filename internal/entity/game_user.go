package entity

import (
	"database/sql"

	"github.com/questx-lab/backend/pkg/enum"
)

type DirectionType string

var (
	Left  = enum.New(DirectionType("left"))
	Right = enum.New(DirectionType("right"))
	Up    = enum.New(DirectionType("up"))
	Down  = enum.New(DirectionType("down"))
)

type GameUser struct {
	RoomID string   `gorm:"primaryKey"`
	Room   GameRoom `gorm:"foreignKey:RoomID"`

	UserID string `gorm:"primaryKey"`
	User   User   `gorm:"foreignKey:UserID"`

	CharacterID string
	Character   GameCharacter `gorm:"foreignKey:CharacterID"`

	Direction   DirectionType
	PositionX   int
	PositionY   int
	ConnectedBy sql.NullString
}
