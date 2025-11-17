//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-16

package zreflect

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

func Dump(obj any) map[string]any {
	visited := map[uintptr]bool{}
	return dumpValue(reflect.ValueOf(obj), visited)
}

func DumpString(obj any) string {
	bf, _ := json.MarshalIndent(Dump(obj), "", "  ")
	return string(bf)
}

func dumpValue(v reflect.Value, visited map[uintptr]bool) map[string]any {
	m := map[string]any{}

	if !v.IsValid() {
		m["Type"] = "nil"
		m["Value"] = nil
		return m
	}

	t := v.Type()
	kind := t.Kind()

	// ---- 基础类型 --------------------------------------------------------
	switch kind {
	case reflect.Bool:
		m["Type"] = "bool"
		m["Value"] = v.Bool()
		return m
	case reflect.String:
		m["Type"] = "string"
		str := v.String()
		m["Len"] = len(str)
		m["Value"] = str
		return m
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		m["Type"] = t.String()
		m["Value"] = v.Int()
		return m
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		m["Type"] = t.String()
		m["Value"] = v.Uint()
		return m
	case reflect.Float32, reflect.Float64:
		m["Type"] = t.String()
		m["Value"] = v.Float()
		return m
	case reflect.Complex64, reflect.Complex128:
		m["Type"] = t.String()
		m["Value"] = v.Complex()
		return m
	}

	// ---- 指针 ------------------------------------------------------------
	if kind == reflect.Ptr {
		m["Type"] = "*" + interfaceToAny(t.Elem().String())

		if v.IsNil() {
			m["Value"] = nil
			return m
		}

		ptr := v.Pointer()
		if visited[ptr] {
			m["Value"] = "...cycle..."
			return m
		}

		visited[ptr] = true
		defer delete(visited, ptr)

		m["Value"] = dumpValue(v.Elem(), visited)
		return m
	}

	// ---- interface --------------------------------------------------------
	if kind == reflect.Interface {
		m["Type"] = "any"
		if v.IsNil() {
			m["Value"] = nil
		} else {
			m["Value"] = dumpValue(v.Elem(), visited)
		}
		return m
	}

	// ---- struct -----------------------------------------------------------
	if kind == reflect.Struct {
		m["Type"] = "struct"
		fields := map[string]any{}
		m["NumField"] = t.NumField()
		m["Fields"] = fields

		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			fv := v.Field(i)

			if !fv.CanInterface() {
				// 未导出字段
				fields[f.Name] = map[string]any{
					"Type":  interfaceToAny(f.Type.String()),
					"Value": "<unexported>",
				}
				continue
			}

			fields[f.Name] = dumpValue(fv, visited)
		}
		return m
	}

	// ---- slice / array ---------------------------------------------------
	if kind == reflect.Slice || kind == reflect.Array {
		m["Type"] = interfaceToAny(t.String())
		size := v.Len()
		items := make([]any, 0, size)
		for i := 0; i < size; i++ {
			item := dumpValue(v.Index(i), visited)
			item["Index"] = i
			items = append(items, item)
		}
		m["Len"] = size
		m["Elem"] = items
		return m
	}

	// ---- map --------------------------------------------------------------
	if kind == reflect.Map {
		m["Type"] = interfaceToAny(t.String())
		m["Len"] = v.Len()

		keys := v.MapKeys()
		// key 排序，保证可读性与可重复性
		sort.Slice(keys, func(i, j int) bool {
			return fmt.Sprintf("%v", keys[i]) < fmt.Sprintf("%v", keys[j])
		})
		elem := make([]map[string]any, 0, len(keys))
		for _, k := range keys {
			val := v.MapIndex(k)
			item := map[string]any{
				"Key":   dumpValue(k, visited),
				"Value": dumpValue(val, visited),
			}
			elem = append(elem, item)
		}
		m["Elem"] = elem

		return m
	}

	// ---- chan / func / other fallback -------------------------------------
	m["Type"] = interfaceToAny(t.String())
	m["Value"] = fmt.Sprintf("%v", v.Interface())
	return m
}

func interfaceToAny(str string) string {
	return strings.ReplaceAll(str, "interface {}", "any")
}
