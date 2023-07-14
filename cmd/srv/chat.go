package main

import (
	"context"

	"github.com/questx-lab/backend/migration"
	gocqlutil "github.com/questx-lab/backend/pkg/cqlutil"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/scylladb/gocqlx/v2"
	"github.com/urfave/cli/v2"
)

func (s *srv) startChat(*cli.Context) error {
	cfg := xcontext.Configs(s.ctx)

	cluster := gocqlutil.CreateCluster(cfg.ScyllaDB.KeySpace, cfg.ScyllaDB.Addr)

	var err error
	s.scyllaDBSession, err = gocqlx.WrapSession(cluster.CreateSession())
	if err != nil {
		panic(err)
	}
	defer s.scyllaDBSession.Close()

	xcontext.Logger(s.ctx).Infof("Connect scylla db successful in addr: %s\n", cfg.ScyllaDB.Addr)

	if err := s.MigrateScyllaDB(s.ctx); err != nil {
		panic(err)
	}
	xcontext.Logger(s.ctx).Infof("Migrate scylla db successful")

	return nil
}

func (s *srv) MigrateScyllaDB(ctx context.Context) error {
	if err := migration.MigrateScyllaDB(ctx, s.scyllaDBSession); err != nil {
		return err
	}

	return nil
}
