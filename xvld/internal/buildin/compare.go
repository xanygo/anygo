package buildin

import (
	"cmp"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/xanygo/anygo/xvld/internal"
)

// minMax 校验数值范围
func minMax(fc *FieldContext, rule Rule) error {
	v := fc.Value()
	if !v.IsValid() {
		return fc.UnsupportedType(rule)
	}
	switch v.Kind() {
	case reflect.String:
		s := v.String()
		if s == "" || compareMinMax(s, rule.Param, rule.Name) {
			return nil
		}
		return fc.Error(rule.Name, rule.Param, internal.ShortErrStr(s))

	case reflect.Array, reflect.Slice:
		// 检查子元素
		for i := 0; i < v.Len(); i++ {
			nc := fc.Child(v.Index(i))
			if err := minMax(nc, rule); err != nil {
				return err
			}
		}
		return nil
	case reflect.Map:
		// 检查子元素
		for _, key := range v.MapKeys() {
			val := v.MapIndex(key)
			nc := fc.Child(val)
			if err := minMax(nc, rule); err != nil {
				return err
			}
		}
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		num, err := strconv.ParseInt(rule.Param, 10, 64)
		if err != nil {
			return fc.InvalidParam(rule, err)
		}
		if compareMinMax(v.Int(), num, rule.Name) {
			return nil
		}
		return fc.Error(rule.Name, num, v.Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		num, err := strconv.ParseUint(rule.Param, 10, 64)
		if err != nil {
			return fc.InvalidParam(rule, err)
		}
		if compareMinMax(v.Uint(), num, rule.Name) {
			return nil
		}
		return fc.Error(rule.Name, num, v.Uint())

	case reflect.Float32, reflect.Float64:
		num, err := strconv.ParseFloat(rule.Param, 64)
		if err != nil {
			return fc.InvalidParam(rule, err)
		}
		if compareMinMax(v.Float(), num, rule.Name) {
			return nil
		}
		return fc.Error(rule.Name, num, v.Float())
	default:
		return fc.UnsupportedType(rule)
	}
}

func compareMinMax[T cmp.Ordered](value, param T, op string) bool {
	switch op {
	case "min", "min-len":
		return value >= param
	case "max", "max-len":
		return value <= param
	case "len":
		return value == param
	}
	return false
}

// validateMinMaxLen 验证长度 (len)
func validateMinMaxLen(fc *FieldContext, rule Rule) error {
	v := fc.Value()
	if !v.IsValid() {
		return fc.UnsupportedType(rule)
	}
	switch v.Kind() {
	case reflect.String:
		length := utf8.RuneCountInString(v.String())
		num, err := strconv.Atoi(rule.Param)
		if err != nil {
			return fc.InvalidParam(rule, err)
		}
		if compareMinMax(length, num, rule.Name) {
			return nil
		}
		return fc.Error(rule.Name, num, length)

	case reflect.Array, reflect.Slice, reflect.Map:
		num, err := strconv.Atoi(rule.Param)
		if err != nil {
			return fc.InvalidParam(rule, err)
		}
		if compareMinMax(v.Len(), num, rule.Name) {
			return nil
		}
		return fc.Error(rule.Name, num, v.Len())
	default:
		return fc.UnsupportedType(rule)
	}
}

// validateNotEqual ne 的校验逻辑，支持 ne=hello,world ,即同时支持 1个或多个关键字
func validateNotEqual(fc *FieldContext, rule Rule) error {
	arr := strings.Split(rule.Param, ",")
	for i := 0; i < len(arr); i++ {
		nr := Rule{Name: rule.Name, Param: strings.TrimSpace(arr[i])}
		if err := validateCompare(fc, nr); err != nil {
			return err
		}
	}
	return nil
}

// validateCompare 比较数值或字符串
func validateCompare(fc *FieldContext, rule Rule) error {
	v := fc.Value()
	if !v.IsValid() {
		return fc.UnsupportedType(rule)
	}
	switch v.Kind() {
	case reflect.String:
		s := v.String()
		if rule.Name == "eq" {
			if s == rule.Param {
				return nil
			}
			return fc.Error(rule.Name, rule.Param, internal.ShortErrStr(s))
		}
		if rule.Name == "ne" {
			if s != rule.Param {
				return nil
			}
			return fc.Error(rule.Name, rule.Param, internal.ShortErrStr(s))
		}
		return fc.UnsupportedType(rule)

	case reflect.Bool:
		if rule.Name == "eq" {
			if strconv.FormatBool(v.Bool()) == rule.Param {
				return nil
			}
			return fc.Error(rule.Name, rule.Param, v.Bool())
		}
		if rule.Name == "ne" {
			if strconv.FormatBool(v.Bool()) != rule.Param {
				return nil
			}
			return fc.Error(rule.Name, rule.Param, v.Bool())
		}
		return fc.UnsupportedType(rule)

	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		num, err := strconv.ParseInt(rule.Param, 10, 64)
		if err != nil {
			return fc.InvalidParam(rule, err)
		}
		return doCompare(fc, rule, num, v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		num, err := strconv.ParseUint(rule.Param, 10, 64)
		if err != nil {
			return fc.InvalidParam(rule, err)
		}
		return doCompare(fc, rule, num, v.Uint())

	case reflect.Float32, reflect.Float64:
		num, err := strconv.ParseFloat(rule.Param, 64)
		if err != nil {
			return fc.InvalidParam(rule, err)
		}
		return doCompare(fc, rule, num, v.Float())
	case reflect.Array, reflect.Slice:
		return internal.CheckSliceElem(fc, rule, validateCompare)
	case reflect.Map:
		return internal.CheckMapValue(fc, rule, validateCompare)
	default:
		return fc.UnsupportedType(rule)
	}
}

func doCompare[T cmp.Ordered](fc *FieldContext, rule Rule, expect T, got T) error {
	switch rule.Name {
	case "eq":
		if got == expect {
			return nil
		}
		return fc.Error(rule.Name, expect, got)
	case "ne":
		if got != got {
			return nil
		}
		return fc.Error(rule.Name, expect, got)
	case "gt":
		if got > expect {
			return nil
		}
		return fc.Error(rule.Name, expect, got)
	case "gte":
		if got >= expect {
			return nil
		}
		return fc.Error(rule.Name, expect, got)
	case "lt":
		if got < expect {
			return nil
		}
		return fc.Error(rule.Name, expect, got)
	case "lte":
		if got <= expect {
			return nil
		}
		return fc.Error(rule.Name, expect, got)
	default:
		// 理论不应该触达
		return fc.UnsupportedType(rule)
	}
}
