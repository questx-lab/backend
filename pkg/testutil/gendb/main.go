package main

import (
	"os"
	"path/filepath"

	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/schollz/sqlite3dump"
)

func main() {
	ctx := testutil.NewMockContext()

	testutil.CreateFixtureContext(ctx)

	f, err := os.Create(filepath.Join("..", testutil.DbDump))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	sqlDb, err := ctx.DB().DB()
	if err != nil {
		panic(err)
	}

	err = sqlite3dump.DumpDB(sqlDb, f)
	if err != nil {
		panic(err)
	}
}
