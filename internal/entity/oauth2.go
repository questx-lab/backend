package entity

type OAuth2 struct {
	UserID string `gorm:"primaryKey"`
	User   User   `gorm:"foreignKey:UserID"`

	Service       string `gorm:"primaryKey"`
	ServiceUserID string `gorm:"unique"`
}

func (OAuth2) TableName() string {
	return "oauth2"
}
