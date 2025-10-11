//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-12-06

package tplfn

import (
	"fmt"
	"html/template"
	"math/rand/v2"
	"reflect"
	"strings"
	"time"
)

func InputSelected(value any) ValueAttr {
	values := valuesMap(value)
	fn := func(current any) template.HTMLAttr {
		var selected string
		cstr := fmt.Sprint(current)
		if _, ok := values[cstr]; ok {
			selected = " selected"
		}
		code := fmt.Sprintf("value=%q%s", template.HTMLEscapeString(cstr), selected)
		return template.HTMLAttr(code)
	}
	return ValueAttrFunc(fn)
}

// valuesMap 用于 InputChecked 和 InputSelected ，当组件支持多选是，可以传入一个 slice
func valuesMap(value any) map[string]struct{} {
	values := make(map[string]struct{}, 1)
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < rv.Len(); i++ {
			key := fmt.Sprint(rv.Index(i).Interface())
			values[key] = struct{}{}
		}
	default:
		key := fmt.Sprint(value)
		values[key] = struct{}{}
	}
	return values
}

func InputChecked(value any) ValueAttr {
	values := valuesMap(value)
	fn := func(current any) template.HTMLAttr {
		var checked string
		cstr := fmt.Sprint(current)
		if _, ok := values[cstr]; ok {
			checked = " checked"
		}
		code := fmt.Sprintf("value=%q%s", template.HTMLEscapeString(cstr), checked)
		return template.HTMLAttr(code)
	}
	return ValueAttrFunc(fn)
}

func EachOfIter(values ...any) Iter[any] {
	var index int
	total := len(values)
	next := func() any {
		val := values[index%total]
		index++
		return val
	}
	return IterNextFunc[any](next)
}

func DateTime(d time.Time) string {
	if d.IsZero() {
		return ""
	}
	return d.Format("2006-01-02 15:04:05")
}

func NowTimeFormat(format string) string {
	return time.Now().Format(format)
}

func RandOfIter(values ...any) Iter[any] {
	next := func() (e any) {
		if len(values) == 0 {
			return e
		}
		return values[rand.IntN(len(values))]
	}
	return IterNextFunc[any](next)
}

func InputObjectName(values ...any) string {
	strs := make([]string, 0, len(values))
	for _, v := range values {
		if v == nil {
			continue
		}
		str := fmt.Sprintf("%v", v)
		if str == "" {
			continue
		}
		strs = append(strs, str)
	}
	if len(strs) <= 1 {
		return strings.Join(strs, "")
	}
	var sb strings.Builder
	sb.WriteString(strs[0])
	for _, value := range strs[1:] {
		sb.WriteString("[")
		sb.WriteString(value)
		sb.WriteString("]")
	}
	return sb.String()
}

// IsOddNumber 判断是否奇数
func IsOddNumber(num any) bool {
	return IsRemainder(num, 2, 1)
}

func IsRemainder(num any, mod int, want int) bool {
	switch vn := num.(type) {
	case int:
		return vn%mod == want
	case int8:
		return int(vn)%mod == want
	case int16:
		return int(vn)%mod == want
	case int32:
		return int(vn)%mod == want
	case int64:
		return int(vn)%mod == want

	case uint:
		return int(vn)%mod == want
	case uint8:
		return int(vn)%mod == want
	case uint16:
		return int(vn)%mod == want
	case uint32:
		return int(vn)%mod == want
	case uint64:
		return int(vn)%mod == want

	case float32:
		return int(vn)%mod == want
	case float64:
		return int(vn)%mod == want
	default:
		return false
	}
}

// IsEvenNumber 判断是否偶数
func IsEvenNumber(num any) bool {
	return IsRemainder(num, 2, 0)
}

func MapKeys(m any) ([]any, error) {
	if m == nil {
		return nil, nil
	}
	v := reflect.ValueOf(m)
	if v.Kind() != reflect.Map {
		return nil, fmt.Errorf("input type is %q not a map", v.Kind().String())
	}

	keys := make([]any, 0, v.Len())

	for _, key := range v.MapKeys() {
		keys = append(keys, key.Interface())
	}

	return keys, nil
}

func Assert(values ...any) (string, error) {
	for idx, v := range values {
		if v == nil {
			return "", fmt.Errorf("required value is nil: param[%d] = %#v", idx, v)
		}
		ok1, ok2 := template.IsTrue(v)
		if ok1 && ok2 {
			continue
		}
		return "", fmt.Errorf("required value is empty: param[%d] = %#v", idx, v)
	}
	return "", nil
}

func OrMap(value any) any {
	ok1, ok2 := template.IsTrue(value)
	if ok1 && ok2 {
		return value
	}
	return map[string]any{}
}

func Join(value any, args ...string) string {
	if value == nil {
		return ""
	}
	var ss []string
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < rv.Len(); i++ {
			item := template.HTMLEscaper(rv.Index(i).Interface())
			ss = append(ss, item)
		}
	default:
		return template.HTMLEscaper(value)
	}
	var sep string
	if len(args) == 1 {
		sep = args[0]
	} else {
		sep = " , "
	}
	return strings.Join(ss, sep)
}

func NL2BR(s string) template.HTML {
	escaped := template.HTMLEscapeString(s)
	withBr := strings.ReplaceAll(escaped, "\n", "<br>")
	return template.HTML(withBr)
}
