package entity

import "database/sql"

type Staff struct {
	Base
	ProjectID sql.NullString
	UserID    sql.NullString
	Role      sql.NullString
	CreatedBy sql.NullString
}
