package utils

import (
	"fmt"
	"reflect"
	"strings"
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
