package zreflect

import (
	"hash/fnv"
	"reflect"
)

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

var staticTypeIDs = map[reflect.Type]uint32{
	reflect.TypeFor[string]():  1,
	reflect.TypeFor[int]():     2,
	reflect.TypeFor[int8]():    3,
	reflect.TypeFor[int16]():   4,
	reflect.TypeFor[int32]():   5,
	reflect.TypeFor[int64]():   6,
	reflect.TypeFor[uint]():    7,
	reflect.TypeFor[uint8]():   8,
	reflect.TypeFor[uint16]():  9,
	reflect.TypeFor[uint32]():  10,
	reflect.TypeFor[uint64]():  11,
	reflect.TypeFor[float32](): 12,
	reflect.TypeFor[float64](): 13,
	reflect.TypeFor[bool]():    14,
	reflect.TypeFor[[]byte]():  15,
}

func TypeID(values ...any) uint32 {
	if len(values) == 0 {
		return 0
	}
	same := true
	types := make([]reflect.Type, len(values))
	for index, item := range values {
		rt := reflect.TypeOf(item)
		types[index] = rt
		if index > 0 && same {
			same = types[0] == rt
		}
	}
	if same {
		if id, ok := staticTypeIDs[types[0]]; ok {
			return id
		}
	}
	return genTypeID(types...)
}

func genTypeID(types ...reflect.Type) uint32 {
	h := fnv.New32a()
	for _, item := range types {
		if item == nil {
			h.Write([]byte("#Nil"))
		} else {
			h.Write([]byte(item.String()))
		}

		h.Write([]byte{0})
	}
	return h.Sum32()
}

func TypeID1[K any]() uint32 {
	rt := reflect.TypeFor[K]()
	if id, ok := staticTypeIDs[rt]; ok {
		return id
	}
	return genTypeID(rt)
}

func TypeID2[K any, V any]() uint32 {
	rt1 := reflect.TypeFor[K]()
	rt2 := reflect.TypeFor[V]()
	if rt1 == rt2 {
		if id, ok := staticTypeIDs[rt1]; ok {
			return id
		}
	}
	return genTypeID(rt1, rt2)
}
