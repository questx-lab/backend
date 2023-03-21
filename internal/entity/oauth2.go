package entity

type OAuth2 struct {
	UserID        string `gorm:"primaryKey"`
	Service       string `gorm:"primaryKey"`
	User          User   `gorm:"foreignKey:UserID"`
	ServiceUserID string
}

func (OAuth2) TableName() string {
	return "oauth2"
}
