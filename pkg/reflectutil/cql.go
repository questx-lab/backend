package reflectutil

import (
	"reflect"
	"regexp"
	"sort"
	"strings"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z]+[a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z]+)")

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func GetColumnNames(i any) []string {
	result := []string{}
	val := reflect.ValueOf(i).Elem()
	for i := 0; i < val.NumField(); i++ {
		name := val.Type().Field(i).Name
		name = ToSnakeCase(name)
		result = append(result, name)
	}
	sort.Strings(result)

	return result
}
