package migration

import (
	"context"
	"database/sql"

	"github.com/questx-lab/backend/pkg/xcontext"
)

// Temporary struct for foreign key contraint.
type User struct{ ID string }
type ClaimedQuest struct{ ID string }

// This struct is cloned from the first version of enity.PayReward. It is used
// to keep the original table structure. If we call CreateTable with
// entity.PayReward instead of this struct, it always creates a table with the
// latest version, and modifying columns of this table in the future may be
// failed.
type PayReward struct {
	// NOTE: Please make sure this is the latest version Base at the time this
	// file is created.
	Base0

	UserID string
	User   User `gorm:"foreignKey:UserID"`

	ClaimedQuestID sql.NullString
	ClaimedQuest   ClaimedQuest `gorm:"foreignKey:ClaimedQuestID"`

	Note    string
	Status  string
	Address string
	Token   string
	Amount  float64

	TxHash string
}

func migrate0008(ctx context.Context) error {
	if err := xcontext.DB(ctx).Migrator().DropTable("transactions"); err != nil {
		return err
	}

	if err := xcontext.DB(ctx).Migrator().CreateTable(&PayReward{}); err != nil {
		return err
	}

	return nil
}
