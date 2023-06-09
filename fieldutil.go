package logiface

import (
	"fmt"
)

// MapFields is a helper function that calls [Builder.Field] or
// [Context.Field], for every element of a map, which must have keys with an
// underlying type of string. The field order is not stable, as it iterates on
// the map, without sorting the keys.
//
// WARNING: The behavior of the [Context.Field] and [Builder.Field] methods may
// change without notice, to facilitate the addition of new field types.
func MapFields[K ~string, V any, R interface {
	Enabled() bool
	Field(key string, val any) R
}](r R, m map[K]V) R {
	if r.Enabled() {
		for k, v := range m {
			r = r.Field(string(k), v)
		}
	}
	return r
}

// ArgFields is a helper function that calls [Builder.Field] or
// [Context.Field], for every element of a (varargs) slice.
// If provided, f will be used to ensure that each key is a string, otherwise,
// if not provided, each key will be converted to a string, using fmt.Sprint.
// Passing an odd number of keys will set the last value to any(nil).
//
// WARNING: The behavior of the [Context.Field] and [Builder.Field] methods may
// change without notice, to facilitate the addition of new field types.
func ArgFields[E any, R interface {
	Enabled() bool
	Field(key string, val any) R
}](r R, f func(key E) (string, bool), l ...E) R {
	if r.Enabled() && len(l) != 0 {
		var (
			key string
			ok  bool
		)
		for i := 0; i < len(l); i += 2 {
			key, ok = argFieldKeyConverter(f, l[i])
			if !ok {
				continue
			}
			if i+1 == len(l) {
				r = r.Field(key, nil)
				break
			}
			r = r.Field(key, l[i+1])
		}
	}
	return r
}

func argFieldKeyConverter[E any](f func(key E) (string, bool), key E) (string, bool) {
	if f == nil {
		return fmt.Sprint(key), true
	}
	return f(key)
}

// SliceArray adds a slice as an array field, to the given [Builder],
// [Context], [ArrayBuilder], [ObjectBuilder], or [Chain].
//
// Note that the key must be empty if the parent parameter is an array.
func SliceArray[E Event, V any, P Parent[E]](parent P, key string, slice []V) P {
	if parent.Enabled() {
		b := ArrayWithKey[E](parent, key)
		for _, v := range slice {
			b.Field(v)
		}
		b.As(key)
	}
	return parent
}

// MapObject adds a map as an object field, to the given [Builder],
// [Context], [ArrayBuilder], [ObjectBuilder], or [Chain].
//
// Note that the key must be empty if the parent parameter is an array.
func MapObject[E Event, K ~string, V any, P Parent[E]](parent P, key string, m map[K]V) P {
	if parent.Enabled() {
		b := ObjectWithKey[E](parent, key)
		for k, v := range m {
			b.Field(string(k), v)
		}
		b.As(key)
	}
	return parent
}
