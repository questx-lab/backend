package entities

import "time"

type User struct {
	ID       string
	Name     string
	Age      int32
	BirthDay time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}
