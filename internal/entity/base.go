package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math/big"
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

type BigInt struct{ *big.Int }

func (b *BigInt) Scan(value any) error {
	switch t := value.(type) {
	case string:
		if _, ok := b.SetString(t, 10); !ok {
			return fmt.Errorf("wrong big.Int type")
		}
	case []byte:
		if _, ok := b.SetString(string(t), 10); !ok {
			return fmt.Errorf("wrong big.Int type")
		}
	default:
		return fmt.Errorf("cannot scan invalid data type %T", value)
	}

	return nil
}

func (b BigInt) Value() (driver.Value, error) {
	return b.String(), nil
}

type NullBigInt struct {
	BigInt BigInt
	Valid  bool
}

func (b *NullBigInt) Scan(value any) error {
	if value == nil {
		b.BigInt, b.Valid = BigInt{}, false
		return nil
	}
	b.Valid = true

	return b.BigInt.Scan(value)
}

// Value implements the driver Valuer interface.
func (b NullBigInt) Value() (driver.Value, error) {
	if !b.Valid {
		return nil, nil
	}

	return b.BigInt.Value()
}
