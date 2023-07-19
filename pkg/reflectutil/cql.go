package reflectutil

import (
	"reflect"
	"sort"

	"github.com/questx-lab/backend/pkg/stringutil"
)

func GetColumnNames(i any) []string {
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
