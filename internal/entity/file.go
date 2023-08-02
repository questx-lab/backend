package entity

type File struct {
	Base
	Mime          string
	Name          string
	CreatedBy     string `gorm:"not null"`
	CreatedByUser User   `gorm:"foreignKey:CreatedBy"`
	Url           string
}

type Bucket string

const (
	Image Bucket = "images"
)
