package entity

import (
	"database/sql"
	"time"
)

type Project struct {
	Base
	CreatedBy      string
	CreatedByUser  User `gorm:"foreignKey:CreatedBy"`
	ReferredBy     sql.NullString
	ReferredByUser User   `gorm:"foreignKey:ReferredBy"`
	Name           string `gorm:"unique"`
	Followers      int
	LogoPictures   Map    // Contains images in different sizes.
	Introduction   []byte `gorm:"type:longtext"`
	Twitter        string
	Discord        string

	WebsiteURL         string
	DevelopmentStage   string
	TeamSize           int
	SharedContentTypes Array[string]
}

type ClaimedReferredProject struct {
	UserID string
	User   User `gorm:"foreignKey:UserID"`

	ProjectID string  `gorm:"unique"`
	Project   Project `gorm:"foreignKey:ProjectID"`

	CreatedAt time.Time
}
