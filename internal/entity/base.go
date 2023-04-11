package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Base struct {
	ID        string `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func MigrateTable(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&User{},
		&OAuth2{},
		&Project{},
		&Quest{},
		&Collaborator{},
		&Category{},
		&ClaimedQuest{},
		&Participant{},
		&APIKey{},
		&RefreshToken{},
		&Achievement{},
	); err != nil {
		return err
	}
	return nil
}

type Array[T any] []T

func (a *Array[T]) Scan(obj any) error {
	switch t := obj.(type) {
	case string:
		return json.Unmarshal([]byte(t), a)
	case []byte:
		return json.Unmarshal(t, a)
	}

	return fmt.Errorf("cannot scan invalid data type %T", obj)
}

func (a Array[T]) Value() (driver.Value, error) {
	return json.Marshal(a)
}
