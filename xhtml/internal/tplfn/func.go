//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-12-06

package tplfn

import (
	"fmt"
	"html/template"
	"math/rand/v2"
	"strings"
	"time"
)

func InputSelected(value any) ValueAttr {
	valuesStr := fmt.Sprint(value)
	fn := func(current any) template.HTMLAttr {
		var selected string
		cstr := fmt.Sprint(current)
		if cstr == valuesStr {
			selected = " selected"
		}
		code := fmt.Sprintf("value=%q%s", template.HTMLEscapeString(cstr), selected)
		return template.HTMLAttr(code)
	}
	return ValueAttrFunc(fn)
}

func InputChecked(value any) ValueAttr {
	valuesStr := fmt.Sprint(value)
	fn := func(current any) template.HTMLAttr {
		var checked string
		cstr := fmt.Sprint(current)
		if cstr == valuesStr {
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
