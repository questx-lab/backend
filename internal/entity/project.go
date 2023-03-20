package entity

import "database/sql"

type Project struct {
	Base
	CreatedBy sql.NullString
	Name      sql.NullString
	Twitter   sql.NullString
	Discord   sql.NullString
	Telegram  sql.NullString
}

func (e *Project) Table() string {
	return "projects"
}
