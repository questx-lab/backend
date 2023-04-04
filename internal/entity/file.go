package entity

type File struct {
	Base
	Mine      string
	Name      string
	CreatedBy string `gorm:"not null"`
	User      User   `gorm:"foreignKey:UserID"`
	UserID    string `gorm:"not null"`
	Url       string
}
