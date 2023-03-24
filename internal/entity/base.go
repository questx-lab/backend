package entity

import (
	"time"

	"gorm.io/gorm"
)

type Base struct {
	ID        string `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Entity interface {
	Table() string
}

func MigrateTable(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&User{},
		&OAuth2{},
		&Project{},
		&Quest{},
		&Collaborator{},
	); err != nil {
		return err
	}
	return nil
}
