package entity

type Category struct {
	Base
	Name        string
	Description string
	CommunityID string    `gorm:"not null"`
	Community   Community `gorm:"foreignKey:CommunityID"`
	CreatedBy   string    `gorm:"not null"`
}
