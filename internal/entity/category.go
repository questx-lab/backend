package entity

import "database/sql"

type Category struct {
	Base
	Name          string
	CommunityID   sql.NullString
	Community     Community `gorm:"foreignKey:CommunityID"`
	CreatedBy     string    `gorm:"not null"`
	CreatedByUser User      `gorm:"foreignKey:CreatedBy"`
}
