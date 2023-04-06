package entity

type APIKey struct {
	ProjectID string  `gorm:"primaryKey"`
	Project   Project `gorm:"foreignKey:ProjectID"`
	Key       string  `gorm:"index"`
}

func (APIKey) TableName() string {
	return "api_keys"
}
