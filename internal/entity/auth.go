package entity

import "time"

type OAuth2 struct {
	UserID string `gorm:"primaryKey"`
	User   User   `gorm:"foreignKey:UserID"`

	Service         string `gorm:"primaryKey"`
	ServiceUserID   string `gorm:"unique"`
	ServiceUsername string
}

func (OAuth2) TableName() string {
	return "oauth2"
}

type RefreshToken struct {
	UserID string
	User   User `gorm:"foreignKey:UserID"`

	Family     string `gorm:"unique;index,unique"`
	Counter    uint64
	Expiration time.Time
}
