package zreflect

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/xanygo/anygo/internal/zerror"
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

// RangeMap 遍历任意类型的 map
//
//   - 只有 map 数据的 key  和 value 的实际类型和 传入的类型完全匹配，才会触发回调
//   - 使用 rv,ok := value.(Type) 方式断言 key 和 value
//   - 不确定的类型，可以使用 any 代替，如只关注 key 的类型是 string，可以使用:Range[string,any](m,func(key string,value any)bool)
//   - 传入的数据可以是任意类型；使用了反射
func RangeMap[K comparable, V any](m any, fn func(key K, val V) error) error {
	if m == nil {
		return nil
	}
	rv := reflect.ValueOf(m)
	if !rv.IsValid() || rv.Kind() != reflect.Map {
		return fmt.Errorf("invalid type, not map: %v", m)
	}
	for _, key := range rv.MapKeys() {
		k, ok := key.Interface().(K)
		if !ok {
			continue
		}
		val, ok := rv.MapIndex(key).Interface().(V)
		if !ok {
			continue
		}
		err := fn(k, val)
		if err == nil {
			continue
		}
		if errors.Is(err, zerror.ErrBreak) {
			return nil
		}
		return err
	}
	return nil
}
