package entity

import "github.com/questx-lab/backend/pkg/enum"

type CollaboratorRole string

var (
	Reviewer = enum.New(CollaboratorRole("reviewer"))
	Owner    = enum.New(CollaboratorRole("owner"))
	Editor   = enum.New(CollaboratorRole("editor"))
)

var ReviewGroup = []CollaboratorRole{Owner, Editor, Reviewer}
var AdminGroup = []CollaboratorRole{Owner, Editor}

type Collaborator struct {
	UserID        string    `gorm:"primaryKey"`
	User          User      `gorm:"foreignKey:UserID"`
	CommunityID   string    `gorm:"primaryKey"`
	Community     Community `gorm:"foreignKey:CommunityID"`
	Role          CollaboratorRole
	CreatedBy     string
	CreatedByUser User `gorm:"foreignKey:CreatedBy"`
}

func (c *Collaborator) TableName() string {
	return "collaborators"
}
