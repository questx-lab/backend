package entity

import "database/sql"

type Category struct {
	Base
	CreatedBy sql.NullString
	ProjectID sql.NullString
}

func (e *Category) Table() string {
	return "categories"
}
