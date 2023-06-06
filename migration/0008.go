package migration

import (
	"context"
	"database/sql"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

// This struct is cloned from the first version of enity.PayReward. It is used
// to keep the original table structure. If we call CreateTable with
// entity.PayReward instead of this struct, it always creates a table with the
// latest version, and modifying columns of this table in the future may be
// failed.
// NOTE: DO NOT MODIFY THIS STRUCT EVEN IF THE ORIGINAL ONE IS MODIFIED.
type PayReward struct {
	entity.Base

	UserID string
	User   entity.User `gorm:"foreignKey:UserID"`

	ClaimedQuestID sql.NullString
	ClaimedQuest   entity.ClaimedQuest `gorm:"foreignKey:ClaimedQuestID"`

	// Note contains the reason of this transaction in case of not come from a
	// claimed quest.
	Note    string
	Status  entity.PayRewardStatusType
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
