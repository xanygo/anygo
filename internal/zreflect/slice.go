package zreflect

import (
	"reflect"
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
