package enum

import (
	"fmt"
	"reflect"
)

var enumManager = map[string]any{}

type enum[T comparable] struct {
	toEnum map[string]T
}

func New[T comparable](value T) T {
	v := reflect.ValueOf(value)
	t := v.Type()
	if _, ok := enumManager[t.Name()]; !ok {
		enumManager[t.Name()] = enum[T]{toEnum: make(map[string]T)}
	}

	enumManager[t.Name()].(enum[T]).toEnum[v.String()] = value
	return value
}

func ToEnum[T comparable](s string) (T, error) {
	var defaultT T
	e, ok := enumManager[reflect.TypeOf(defaultT).Name()]
	if !ok {
		return defaultT, fmt.Errorf("not found enum type %T", defaultT)
	}

	t, ok := e.(enum[T]).toEnum[s]
	if !ok {
		return defaultT, fmt.Errorf("not found value %s in enum %T", s, defaultT)
	}

	return t, nil
}
