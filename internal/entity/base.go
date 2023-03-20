package entity

import "time"

type Base struct {
	ID string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}
