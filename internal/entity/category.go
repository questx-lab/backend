package entity

import "database/sql"

type Category struct {
	Base
	CreatedBy sql.NullString
	ProjectID sql.NullString
}
