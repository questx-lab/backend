package cqlutil

import (
	"reflect"
	"sort"

	"github.com/questx-lab/backend/pkg/stringutil"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"
	"github.com/scylladb/gocqlx/table"
)

func Insert(session *gocql.Session,
	tbl *table.Table,
	data interface{},
) error {
	insertStmt, insertNames := tbl.Insert()
	err := gocqlx.Query(session.Query(insertStmt),
		insertNames).BindStruct(data).ExecRelease()
	if err != nil {
		return err
	}

	return nil
}

func Delete(session *gocql.Session,
	tbl *table.Table,
	data interface{},
) error {
	stmt, names := tbl.Delete()
	err := gocqlx.Query(session.Query(stmt),
		names).BindStruct(data).ExecRelease()
	if err != nil {
		return err
	}

	return nil
}

func Select[T any](session *gocql.Session,
	tbl *table.Table,
	filter T,
) ([]T, error) {
	var result []T
	metadata := tbl.Metadata()
	stmt, names := qb.Select(metadata.Name).Columns(metadata.Columns...).Where().ToCql()
	err := gocqlx.Query(session.Query(stmt),
		names).SelectRelease(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func GetTableNames(i interface{}) []string {
	result := []string{}
	val := reflect.ValueOf(i).Elem()
	for i := 0; i < val.NumField(); i++ {
		name := val.Type().Field(i).Name
		name = stringutil.ToSnakeCase(name)
		result = append(result, name)
	}
	sort.Strings(result)

	return result
}
