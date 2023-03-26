package main

import (
	"os"
	"path/filepath"

	"github.com/questx-lab/backend/pkg/testutil"
)

const (
	DumpFile = "testdb.dump"
)

func main() {
	db := testutil.CreateFixtureDb()

	f, err := os.Create(filepath.Join("..", testutil.DbDump))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	sqlDb, err := db.DB()

	err = DumpDB(sqlDb, f)
	if err != nil {
		panic(err)
	}
}
