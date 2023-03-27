package entity

type Role string

const (
	Reviewer Role = "Reviewer"
	Owner    Role = "Owner"
	Editor   Role = "Editor"
)

type Collaborator struct {
	Base
	UserID    string  `gorm:"not null"`
	ProjectID string  `gorm:"not null"`
	Project   Project `gorm:"foreignKey:ProjectID"`
	User      User    `gorm:"foreignKey:UserID"`
	Role      Role
	CreatedBy string `gorm:"not null"`
}

var Roles = []Role{Reviewer, Owner, Editor}
