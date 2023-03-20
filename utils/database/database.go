package database

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/questx-lab/backend/internal/entity"
)

func FieldMap(e interface{}) ([]string, []interface{}) {
	var fieldNames []string
	var fieldValues []interface{}

	v := reflect.ValueOf(&e).Elem()

	for i := 0; i < v.NumField(); i++ {

		field := v.Type().Field(i)
		fieldName := field.Tag.Get("db")
		fieldValue := v.Field(i).Addr().Interface()

		fieldNames = append(fieldNames, fieldName)
		fieldValues = append(fieldValues, fieldValue)
	}

	return fieldNames, fieldValues
}

func GeneratePlaceHolder(n int) string {
	var result []string

	for i := 1; i <= n; i++ {
		result = append(result, fmt.Sprintf("$%d", i))
	}

	return strings.Join(result, ", ")
}

func Insert(ctx context.Context, db *sql.DB, e entity.Entity) error {

	fields, values := FieldMap(e)
	tableName := e.Table()
	fieldsStr := strings.Join(fields, ", ")
	placeHolder := GeneratePlaceHolder(len(fields))

	stmt := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES(%s)`,
		tableName,
		fieldsStr,
		placeHolder,
	)
	if _, err := db.ExecContext(ctx, stmt, values...); err != nil {
		return fmt.Errorf("unable to insert: %w", err)
	}
	return nil
}
