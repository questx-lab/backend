package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/questx-lab/backend/pkg/logger"
	"gorm.io/gorm"
)

type Base struct {
	ID        string `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func MigrateTable(db *gorm.DB) error {
	err := db.AutoMigrate(
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
		&UserAggregate{},
		&File{},
		&Badge{},
		&GameMap{},
		&GameRoom{},
		&GameUser{},
	)
	if err != nil {
		return err
	}

	return nil
}

func MigrateMySQL(db *gorm.DB, logger logger.Logger) {
	err := db.Exec("CREATE FULLTEXT INDEX `search_project_idx` ON `projects`(`name`,`introduction`)").Error
	if err != nil {
		logger.Warnf("Cannot create search_project_idx: %v", err)
	}

	err = db.Exec("CREATE FULLTEXT INDEX `search_quest_idx` ON `quests`(`title`,`description`)").Error
	if err != nil {
		logger.Warnf("Cannot create search_quest_idx: %v", err)
	}
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
