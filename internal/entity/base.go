package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Base struct {
	ID        string         `redis:"id" gorm:"primarykey"`
	CreatedAt time.Time      `redis:"created_at"`
	UpdatedAt time.Time      `redis:"updated_at"`
	DeletedAt gorm.DeletedAt `redis:"deleted_at" gorm:"index"`
}

type SnowFlakeBase struct {
	ID        int64 `gorm:"primaryKey"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
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

func (m *Map) Scan(value any) error {
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
