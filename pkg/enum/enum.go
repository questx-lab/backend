package enum

import "reflect"

var enumManager = map[string]any{}

type enum[T comparable] struct {
	toString map[T]string
	toEnum   map[string]T
}

func New[T comparable](value T, s string) T {
	t := reflect.TypeOf(value)
	if _, ok := enumManager[t.Name()]; !ok {
		enumManager[t.Name()] = enum[T]{
			toEnum:   make(map[string]T),
			toString: make(map[T]string),
		}
	}

	enumManager[t.Name()].(enum[T]).toString[value] = s
	enumManager[t.Name()].(enum[T]).toEnum[s] = value
	return value
}

func ToString[T comparable](v T) string {
	e, ok := enumManager[reflect.TypeOf(v).Name()]
	if !ok {
		return ""
	}

	s, ok := e.(enum[T]).toString[v]
	if !ok {
		return ""
	}

	return s
}

func ToEnum[T comparable](s string) T {
	var defaultT T
	e, ok := enumManager[reflect.TypeOf(defaultT).Name()]
	if !ok {
		return defaultT
	}

	t, ok := e.(enum[T]).toEnum[s]
	if !ok {
		return defaultT
	}

	return t
}
