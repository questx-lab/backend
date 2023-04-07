package entity

type File struct {
	Base
	Mine      string
	Name      string
	CreatedBy string `gorm:"not null"`
	User      User   `gorm:"foreignKey:UserID"`
	Url       string
}

type Bucket string

const (
	Image Bucket = "images"
)
