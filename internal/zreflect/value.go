package zreflect

import (
	"reflect"
	"strconv"
)

func NumberEqual(a, b reflect.Value) bool {
	ak, bk := a.Kind(), b.Kind()

	// int vs int
	if IsIntKind(ak) && IsIntKind(bk) {
		return a.Int() == b.Int()
	}

	// uint vs uint
	if IsUintKind(ak) && IsUintKind(bk) {
		return a.Uint() == b.Uint()
	}

	// int vs uint
	if IsIntKind(ak) && IsUintKind(bk) {
		ai := a.Int()
		if ai < 0 {
			return false
		}
		return uint64(ai) == b.Uint()
	}

	// uint vs int
	if IsUintKind(ak) && IsIntKind(bk) {
		bi := b.Int()
		if bi < 0 {
			return false
		}
		return a.Uint() == uint64(bi)
	}

	return false
}

// IndirectValue 解引用指针和接口
func IndirectValue(v reflect.Value) reflect.Value {
	for {
		if !v.IsValid() {
			return reflect.Value{}
		}

		switch v.Kind() {
		case reflect.Pointer, reflect.Interface:
			if v.IsNil() {
				return reflect.Value{}
			}
			v = v.Elem()
		default:
			return v
		}
	}
}

// IsEmptyValue 判断空值
func IsEmptyValue(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}

	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0

	case reflect.Bool:
		return !v.Bool()

	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		return v.Int() == 0

	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0

	case reflect.Float32, reflect.Float64:
		return v.Float() == 0

	case reflect.Array, reflect.Slice, reflect.Map:
		return v.Len() == 0

	case reflect.Struct:
		return v.IsZero()
	}

	return false
}

func RangeStructValue(v reflect.Value, fn func(field reflect.StructField, value reflect.Value) bool) {
	v = IndirectValue(v)
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		if !fn(t.Field(i), v.Field(i)) {
			return
		}
	}
}

// StringEqualValue 比对
func StringEqualValue(v reflect.Value, param string) bool {
	if !v.IsValid() {
		return false
	}

	switch v.Kind() {
	case reflect.String:
		return v.String() == param

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10) == param

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10) == param

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(param, 64)
		return err == nil && f == v.Float()
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()) == param
	default:
		return false
	}
}
