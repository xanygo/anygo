package zreflect

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/xanygo/anygo/internal/zerror"
)

func SliceContains(slice any, value any) bool {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return false
	}
	vv := reflect.ValueOf(value)

	for i := 0; i < rv.Len(); i++ {
		sv := rv.Index(i)

		if sv.Type() == vv.Type() {
			if reflect.DeepEqual(sv.Interface(), value) {
				return true
			}
			continue
		}

		if NumberEqual(sv, vv) {
			return true
		}

		if reflect.DeepEqual(sv.Interface(), value) {
			return true
		}
	}

	return false
}

// RangeSlice 遍历任意类型的 slice 、Array
//
//   - 只有 slice 数据的实际类型和 传入的类型完全匹配，才会触发回调
//   - 使用 rv,ok := value.(Type) 方式断言
//   - 不确定的类型，可以使用 any 代替
//   - 传入的数据可以是任意类型；使用了反射
func RangeSlice[T any](obj any, fn func(item T) error) error {
	if obj == nil {
		return nil
	}
	rv := reflect.ValueOf(obj)
	if !rv.IsValid() {
		return fmt.Errorf("rangeSlice with invalid value %T", obj)
	}
	switch rv.Kind() {
	case reflect.Array, reflect.Slice:
	default:
		return fmt.Errorf("%w, expect array/slice, got %T", zerror.ErrInvalidType, obj)
	}
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i).Interface()
		val, ok := elem.(T)
		if !ok {
			continue
		}
		err := fn(val)
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
