package migration

import (
	"context"

	"github.com/questx-lab/backend/pkg/xcontext"
)

func migrate0008(ctx context.Context) error {
	if err := xcontext.DB(ctx).Migrator().DropTable("transactions"); err != nil {
		return err
	}

	return nil
}
