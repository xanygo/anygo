package zreflect

import "reflect"

// ToPlainObject 将对象转换为普通 Go 对象。
//
// Struct 会转换为 map[string]any，并使用 Go 原始字段名，不受 `json` struct tag 的影响。Slice、Array、Map、Pointer
// 等复合类型会递归转换。
//
// 返回值可直接用于 encoding/json.Marshal，此时生成的 JSON 将使用 Go 字段名，而不是 `json` tag 中定义的名称。
func ToPlainObject(obj any) any {
	if obj == nil {
		return nil
	}
	return convertValue(reflect.ValueOf(obj))
}

func convertValue(v reflect.Value) any {
	if !v.IsValid() {
		return nil
	}

	// interface
	if v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil
		}
		return convertValue(v.Elem())
	}

	// pointer
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		return convertValue(v.Elem())
	}

	switch v.Kind() {

	case reflect.Struct:
		return convertStruct(v)

	case reflect.Slice, reflect.Array:
		vk := v.Type().Elem().Kind()
		if IsBasicKind(vk) || IsComplexKind(vk) {
			return v.Interface()
		}
		arr := make([]any, v.Len())
		for i := 0; i < v.Len(); i++ {
			arr[i] = convertValue(v.Index(i))
		}
		return arr

	case reflect.Map:
		m := make(map[string]any)
		iter := v.MapRange()
		for iter.Next() {
			key := iter.Key()

			// json object key 必须 string
			if key.Kind() == reflect.String {
				m[key.String()] = convertValue(iter.Value())
			}
		}

		return m
	default:
		return v.Interface()
	}
}

func convertStruct(v reflect.Value) map[string]any {
	t := v.Type()
	m := make(map[string]any)
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		// 跳过非导出字段
		if sf.PkgPath != "" {
			continue
		}
		m[sf.Name] = convertValue(v.Field(i))
	}
	return m
}
