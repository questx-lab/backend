package entity

type Project struct {
	Base
	CreatedBy    string `gorm:"not null"`
	Name         string `gorm:"unique"`
	LogoPictures Map    // Contains images in different sizes.
	Introduction []byte `gorm:"type:longtext"`
	Twitter      string
	Discord      string

	WebsiteURL         string
	DevelopmentStage   string
	TeamSize           int
	SharedContentTypes Array[string]
}
