package entity

import (
	"time"
)

type Project struct {
	ID        string `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time `gorm:"index"`

	CreatedBy string
	Name      string
	Twitter   string
	Discord   string
	Telegram  string
}

func (e *Project) Table() string {
	return "projects"
}
