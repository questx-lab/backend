package cqlutil

import (
	"reflect"
	"sort"

	"github.com/questx-lab/backend/pkg/stringutil"

	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/qb"
	"github.com/scylladb/gocqlx/v2/table"
)

func Insert(session gocqlx.Session,
	tbl *table.Table,
	data interface{},
) error {
	stmt, names := tbl.Insert()
	err := gocqlx.Session.Query(session, stmt,
		names).BindStruct(data).ExecRelease()
	if err != nil {
		return err
	}

	return nil
}

func Delete(session gocqlx.Session,
	tbl *table.Table,
	data interface{},
) error {
	stmt, names := tbl.Delete()
	err := gocqlx.Session.Query(session, stmt, names).BindStruct(data).ExecRelease()
	if err != nil {
		return err
	}

	return nil
}

func Select[T any](session gocqlx.Session,
	tbl *table.Table,
	filter T,
	limit int64,
	w ...qb.Cmp,
) ([]T, error) {
	var result []T
	metadata := tbl.Metadata()

	stmt, names := qb.Select(metadata.Name).Columns(metadata.Columns...).Where(w...).Limit(uint(limit)).ToCql()
	err := gocqlx.Session.Query(session, stmt,
		names).BindStruct(filter).SelectRelease(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func Update(session gocqlx.Session,
	tbl *table.Table,
	data interface{},
) error {
	stmt, names := tbl.Update()
	err := gocqlx.Session.Query(session, stmt, names).BindStruct(data).ExecRelease()
	if err != nil {
		return err
	}

	return nil
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
