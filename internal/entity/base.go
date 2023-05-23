package entity

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type Base struct {
	ID        string `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func MigrateTable(ctx context.Context) error {
	err := xcontext.DB(ctx).AutoMigrate(
		&User{},
		&OAuth2{},
		&Community{},
		&Quest{},
		&Collaborator{},
		&Category{},
		&ClaimedQuest{},
		&Follower{},
		&APIKey{},
		&RefreshToken{},
		&UserAggregate{},
		&File{},
		&Badge{},
		&GameMap{},
		&GameRoom{},
		&GameUser{},
		&Transaction{},
	)
	if err != nil {
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

type Map map[string]any

func (m *Map) Scan(value interface{}) error {
	switch t := value.(type) {
	case string:
		return json.Unmarshal([]byte(t), m)
	case []byte:
		return json.Unmarshal(t, m)
	default:
		return fmt.Errorf("cannot scan invalid data type %T", value)
	}
}

func (m Map) Value() (driver.Value, error) {
	return json.Marshal(m)
}
