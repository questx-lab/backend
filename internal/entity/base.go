package entity

import (
	"database/sql"
)

type Base struct {
	ID sql.NullString

	CreatedAt sql.NullTime
	UpdatedAt sql.NullTime
	DeletedAt sql.NullTime
}

type Entity interface {
	Table() string
}
