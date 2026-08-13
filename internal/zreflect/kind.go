package zreflect

import "reflect"

func IsIntKind(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	default:
		return false
	}
}

func IsUintKind(k reflect.Kind) bool {
	switch k {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true
	default:
		return false
	}
}

func IsBasicKind(k reflect.Kind) bool {
	switch k {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		return true
	default:
		return false
	}
}

func IsComplexKind(k reflect.Kind) bool {
	switch k {
	case reflect.Complex64, reflect.Complex128:
		return true
	default:
		return false
	}
}

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

func IsBytesArray(rt reflect.Type) bool {
	return rt.Kind() == reflect.Array && rt.Elem().Kind() == reflect.Uint8
}
