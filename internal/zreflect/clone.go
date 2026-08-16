package zreflect

import (
	"fmt"
	"reflect"
	"time"
)

type cloneVisit struct {
	typ reflect.Type
	ptr uintptr
}

type CloneFunc func(v reflect.Value) reflect.Value

var cloneRegistry = make(map[reflect.Type]CloneFunc)

func init() {
	RegisterCloneValue(reflect.TypeOf(time.Time{}), func(v reflect.Value) reflect.Value {
		return v
	})
}

// RegisterClone 注册自定义拷贝实现
//
// Example:
//
//	RegisterClone(func(v MyType) MyType {
//	    return v.Clone()
//	})
func RegisterClone[T any](fn func(T) T) {
	typ := reflect.TypeOf((*T)(nil)).Elem()

	RegisterCloneValue(typ, func(v reflect.Value) reflect.Value {
		return reflect.ValueOf(fn(v.Interface().(T)))
	})
}

func RegisterCloneValue(typ reflect.Type, fn CloneFunc) {
	if typ == nil {
		panic("clone type cannot be nil")
	}
	if fn == nil {
		panic(fmt.Sprintf("%s: clone function cannot be nil", typ.String()))
	}
	cloneRegistry[typ] = fn
}

// Clone 深拷贝.
//
// 策略:
//
//  1. 优先使用已注册的专用 Clone 函数。
//  2. 如果存在 Clone() 方法，且其返回类型与当前值类型完全一致，则调用该方法。
//  3. 对指针、接口、结构体、切片、数组和 Map 进行递归克隆。
//  4. 对于结构体，其导出字段进行深度克隆，未导出字段保持浅复制。
//  5. 标量类型直接复制。
//  6. 支持循环引用。
func Clone[T any](v T) T {
	rv := reflect.ValueOf(v)

	if !rv.IsValid() {
		var zero T
		return zero
	}

	cv := deepClone(rv, make(map[cloneVisit]reflect.Value))
	return cv.Interface().(T)
}

func deepClone(v reflect.Value, visited map[cloneVisit]reflect.Value) reflect.Value {
	if !v.IsValid() {
		return v
	}

	// 1. 使用已注册的专用 Clone 函数
	if fn, ok := cloneRegistry[v.Type()]; ok {
		return fn(v)
	}

	// 2. 尝试使用类型存在的 Clone()T 方法
	if cloned, ok := callClone(v); ok {
		return cloned
	}

	switch v.Kind() {
	case reflect.Pointer:
		return clonePointer(v, visited)

	case reflect.Interface:
		return cloneInterface(v, visited)

	case reflect.Struct:
		return cloneStruct(v, visited)

	case reflect.Slice:
		return cloneSlice(v, visited)

	case reflect.Array:
		return cloneArray(v, visited)

	case reflect.Map:
		return cloneMap(v, visited)

	default:
		// bool, int, uint, float, complex, string, chan,
		// func, unsafe.Pointer, etc.
		//
		// 直接返回值
		return v
	}
}

// 类型实现了 Clone()T 方法
func callClone(v reflect.Value) (reflect.Value, bool) {
	method := v.MethodByName("Clone")
	if !method.IsValid() {
		return reflect.Value{}, false
	}

	mt := method.Type()

	if mt.NumIn() != 0 || mt.NumOut() != 1 {
		return reflect.Value{}, false
	}

	if mt.Out(0) != v.Type() {
		return reflect.Value{}, false
	}

	result := method.Call(nil)
	return result[0], true
}

func clonePointer(v reflect.Value, visited map[cloneVisit]reflect.Value) reflect.Value {
	if v.IsNil() {
		return reflect.Zero(v.Type())
	}

	key := cloneVisit{
		typ: v.Type(),
		ptr: v.Pointer(),
	}

	if cached, ok := visited[key]; ok {
		return cached
	}

	out := reflect.New(v.Type().Elem())
	visited[key] = out

	out.Elem().Set(deepClone(v.Elem(), visited))

	return out
}

func cloneInterface(v reflect.Value, visited map[cloneVisit]reflect.Value) reflect.Value {
	if v.IsNil() {
		return reflect.Zero(v.Type())
	}

	elem := deepClone(v.Elem(), visited)

	out := reflect.New(v.Type()).Elem()
	out.Set(elem)

	return out
}

func cloneStruct(v reflect.Value, visited map[cloneVisit]reflect.Value) reflect.Value {
	out := reflect.New(v.Type()).Elem()
	// 先浅拷贝
	out.Set(v)

	// 对字段深拷贝
	for i := 0; i < v.NumField(); i++ {
		src := v.Field(i)
		dst := out.Field(i)

		if !dst.CanSet() {
			continue
		}

		dst.Set(deepClone(src, visited))
	}

	return out
}

func cloneSlice(v reflect.Value, visited map[cloneVisit]reflect.Value) reflect.Value {
	if v.IsNil() {
		return reflect.Zero(v.Type())
	}

	key := cloneVisit{
		typ: v.Type(),
		ptr: v.Pointer(),
	}

	if cached, ok := visited[key]; ok {
		return cached
	}

	out := reflect.MakeSlice(v.Type(), v.Len(), v.Len())

	// Register before recursively cloning elements.
	visited[key] = out

	for i := 0; i < v.Len(); i++ {
		out.Index(i).Set(deepClone(v.Index(i), visited))
	}

	return out
}

func cloneArray(v reflect.Value, visited map[cloneVisit]reflect.Value) reflect.Value {
	out := reflect.New(v.Type()).Elem()

	for i := 0; i < v.Len(); i++ {
		out.Index(i).Set(deepClone(v.Index(i), visited))
	}

	return out
}

func cloneMap(v reflect.Value, visited map[cloneVisit]reflect.Value) reflect.Value {
	if v.IsNil() {
		return reflect.Zero(v.Type())
	}

	key := cloneVisit{
		typ: v.Type(),
		ptr: v.Pointer(),
	}

	if cached, ok := visited[key]; ok {
		return cached
	}

	out := reflect.MakeMapWithSize(v.Type(), v.Len())

	visited[key] = out

	iter := v.MapRange()
	for iter.Next() {
		newKey := deepClone(iter.Key(), visited)
		value := deepClone(iter.Value(), visited)

		out.SetMapIndex(newKey, value)
	}

	return out
}
