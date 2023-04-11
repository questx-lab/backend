package entity

type File struct {
	Base
	Mime      string
	Name      string
	CreatedBy string `gorm:"not null"`
	UserID    string
	User      User `gorm:"foreignKey:UserID"`
	Url       string
}

type Bucket string

const (
	Image Bucket = "images"
)
