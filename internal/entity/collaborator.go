package entity

import "github.com/questx-lab/backend/pkg/enum"

type Role string

var (
	Reviewer = enum.New(Role("reviewer"))
	Owner    = enum.New(Role("owner"))
	Editor   = enum.New(Role("editor"))
)

var ReviewGroup = []Role{Owner, Editor, Reviewer}
var AdminGroup = []Role{Owner, Editor}

type Collaborator struct {
	Base
	UserID    string  `gorm:"not null"`
	ProjectID string  `gorm:"not null"`
	Project   Project `gorm:"foreignKey:ProjectID"`
	User      User    `gorm:"foreignKey:UserID"`
	Role      Role
	CreatedBy string `gorm:"not null"`
}

func (c *Collaborator) TableName() string {
	return "collaborators"
}
