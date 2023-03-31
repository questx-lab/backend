package reflectutil

import "reflect"

func PartialEqual[T any](a, b T) bool {
	va := reflect.ValueOf(a).Elem()
	vb := reflect.ValueOf(b).Elem()

	for _, field := range reflect.VisibleFields(va.Type()) {
		fieldA := va.FieldByIndex(field.Index)
		fieldB := vb.FieldByIndex(field.Index)

		if !field.IsExported() || fieldA.IsZero() {
			continue
		}

		if fieldA.Interface() != fieldB.Interface() {
			return false
		}
	}
	return true
}
