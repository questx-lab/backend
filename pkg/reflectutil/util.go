package reflectutil

import (
	"reflect"

	"golang.org/x/exp/slices"
)

var numberKinds = []reflect.Kind{
	reflect.Float32, reflect.Float64,
	reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
	reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
}

var supportedDirectCompareKinds = []reflect.Kind{
	reflect.String, reflect.Bool,
	reflect.Float32, reflect.Float64,
	reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
	reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
}

func PartialEqual[T any](a, b T) bool {
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)
	if va.Kind() == reflect.Pointer || va.Kind() == reflect.Interface {
		va = va.Elem()
		vb = vb.Elem()
	}

	if va.IsZero() {
		return true
	}

	if slices.Contains(supportedDirectCompareKinds, va.Kind()) {
		if isNumber(va) {
			if !isNumber(vb) {
				return false
			}

			return getNumber(va) == getNumber(vb)
		}
		return va.Interface() == vb.Interface()
	}

	if va.Kind() == reflect.Map {
		for _, k := range va.MapKeys() {
			fieldA := va.MapIndex(k)
			fieldB := vb.MapIndex(k)
			if !PartialEqual(fieldA.Interface(), fieldB.Interface()) {
				return false
			}
		}

		return true
	}

	for _, field := range reflect.VisibleFields(va.Type()) {
		fieldA := va.FieldByIndex(field.Index)
		fieldB := vb.FieldByIndex(field.Index)

		if !field.IsExported() || fieldA.IsZero() {
			continue
		}

		if field.Type.Kind() == reflect.Array || field.Type.Kind() == reflect.Slice {
			if fieldA.Len() != fieldB.Len() {
				return false
			}

			for i := 0; i < fieldA.Len(); i++ {
				ai := fieldA.Index(i).Interface()
				bi := fieldB.Index(i).Interface()
				if !PartialEqual(ai, bi) {
					return false
				}
			}
		} else {
			if !PartialEqual(fieldA.Interface(), fieldB.Interface()) {
				return false
			}
		}
	}

	return true
}

func isNumber(v reflect.Value) bool {
	return slices.Contains(numberKinds, v.Kind())
}

func getNumber(v reflect.Value) float64 {
	switch v.Kind() {
	case reflect.Float64:
		return v.Interface().(float64)
	case reflect.Float32:
		return float64(v.Interface().(float32))
	case reflect.Int:
		return float64(v.Interface().(int))
	case reflect.Int8:
		return float64(v.Interface().(int8))
	case reflect.Int16:
		return float64(v.Interface().(int16))
	case reflect.Int32:
		return float64(v.Interface().(int32))
	case reflect.Int64:
		return float64(v.Interface().(int64))
	case reflect.Uint:
		return float64(v.Interface().(uint))
	case reflect.Uint8:
		return float64(v.Interface().(uint8))
	case reflect.Uint16:
		return float64(v.Interface().(uint16))
	case reflect.Uint32:
		return float64(v.Interface().(uint32))
	case reflect.Uint64:
		return float64(v.Interface().(uint64))
	}

	panic("invalid type")
}
