package reflectutil

import "reflect"

func PartialEqual[T any](a T, b T) bool {
	va := reflect.ValueOf(a).Elem()
	vb := reflect.ValueOf(b).Elem()

	for i := 0; i < va.NumField(); i++ {
		fieldA := va.Field(i)
		fieldB := vb.Field(i)

		if fieldA.IsZero() {
			continue
		}
		if fieldA.Interface() != fieldB.Interface() {

			return false
		}
	}
	return true
}
