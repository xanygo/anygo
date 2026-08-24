package zreflect

import (
	"errors"
	"fmt"
	"reflect"
)

func MapHasKey(m any, key any) (bool, error) {
	rv := reflect.ValueOf(m)

	if !rv.IsValid() {
		return false, fmt.Errorf("is nil: %#v", m)
	}

	if rv.Kind() != reflect.Map {
		return false, fmt.Errorf("expected map, got %s", rv.Kind())
	}

	kv := reflect.ValueOf(key)
	if !kv.IsValid() {
		return false, errors.New("key is nil")
	}

	if rv.IsNil() {
		return false, nil
	}

	keyType := rv.Type().Key()

	if !kv.Type().AssignableTo(keyType) {
		if !kv.Type().ConvertibleTo(keyType) {
			return false, fmt.Errorf("key type %s cannot be used as %s", kv.Type(), keyType)
		}

		kv = kv.Convert(keyType)
	}

	return rv.MapIndex(kv).IsValid(), nil
}
