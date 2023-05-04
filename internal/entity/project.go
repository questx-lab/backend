package entity

type Project struct {
	Base
	CreatedBy    string `gorm:"not null"`
	Name         string
	Introduction []byte
	Twitter      string
	Discord      string
}
