package main

import (
	"fmt"

	"github.com/questx-lab/backend/migration"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/urfave/cli/v2"
)

func (s *srv) startMigrate(cctx *cli.Context) error {
	s.ctx = xcontext.WithDB(s.ctx, s.newDatabase())
	s.migrateDB()

	version := cctx.String("version")
	migrator, ok := migration.Migrators[version]
	if !ok {
		return fmt.Errorf("not found version %s", version)
	}

	if err := migrator(s.ctx); err != nil {
		return err
	}

	return nil
}
