package entity

type APIKey struct {
	CommunityID string    `gorm:"primaryKey"`
	Community   Community `gorm:"foreignKey:CommunityID"`
	Key         string    `gorm:"index"`
}

func (APIKey) TableName() string {
	return "api_keys"
}
