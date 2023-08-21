package entity

import (
	"bytes"
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

type SnowFlakeBase struct {
	ID        int64 `gorm:"primaryKey"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Array[T any] []T

func (a *Array[T]) Scan(obj any) error {
	var body []byte
	switch t := obj.(type) {
	case string:
		body = []byte(t)
	case []byte:
		body = t
	default:
		return fmt.Errorf("cannot scan invalid data type %T", obj)
	}

	d := json.NewDecoder(bytes.NewBuffer(body))
	d.UseNumber()
	return d.Decode(a)
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
