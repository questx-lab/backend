package entity

type Project struct {
	Base
	CreatedBy    string `gorm:"not null"`
	Name         string
	Introduction []byte `gorm:"type:longtext"`
	Twitter      string
	Discord      string
}
