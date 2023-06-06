package migration

import (
	"time"

	"gorm.io/gorm"
)

// Create a new version of Base if the entity.Base has changed.
// NOTE: DO NOT DELETE THIS STRUCT.
type Base0 struct {
	ID        string `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
