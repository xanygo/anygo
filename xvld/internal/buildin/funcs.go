package buildin

import (
	"net"
	"net/url"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"uuid"

	"github.com/xanygo/anygo/internal/zcache"
	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/xstr"
	"github.com/xanygo/anygo/xvld/internal"
)

const domainRegText = `[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)+$`

const mailRegText = `^[a-zA-Z0-9]+([._%+-][a-zA-Z0-9]+)*@` + domainRegText

func required(fc *FieldContext, rule Rule) error {
	if !zreflect.IsEmptyValue(fc.Value()) {
		return nil
	}
	return fc.Error(rule.Name)
}

// oneOf 枚举校验
func oneOf(fc *FieldContext, rule Rule) error {
	v := fc.Value()
	if !v.IsValid() {
		return fc.UnsupportedType(rule)
	}
	arr := xstr.ToStrings(rule.Param, "|")
	if len(arr) == 0 {
		return fc.InvalidParam(rule, internal.ErrParamEmpty)
	}
	switch v.Kind() {
	case reflect.String:
		s := v.String()
		// 若要判断值非空，应配合 required 使用
		if s == "" || slices.Contains(arr, s) {
			return nil
		}
		return fc.Error(rule.Name, rule.Param, internal.ShortErrStr(s))
	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		str := strconv.FormatInt(v.Int(), 10)
		if slices.Contains(arr, str) {
			return nil
		}
		return fc.Error(rule.Name, rule.Param, str)

	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		str := strconv.FormatUint(v.Uint(), 10)
		if slices.Contains(arr, str) {
			return nil
		}
		return fc.Error(rule.Name, rule.Param, str)

	case reflect.Float32, reflect.Float64:
		vf := v.Float()
		for _, s := range arr {
			n, err := strconv.ParseFloat(s, v.Type().Bits())
			if err == nil && vf == n {
				return nil
			}
		}
		return fc.Error(rule.Name, rule.Param, vf)
	case reflect.Bool:
		str := strconv.FormatBool(v.Bool())
		if slices.Contains(arr, str) {
			return nil
		}
		return fc.Error(rule.Name, rule.Param, str)
	case reflect.Array, reflect.Slice:
		return internal.CheckSliceElem(fc, rule, oneOf)
	case reflect.Map:
		return internal.CheckMapValue(fc, rule, oneOf)
	default:
		return fc.UnsupportedType(rule)
	}
}

func isURL(str string, _ Rule) bool {
	u, err := url.ParseRequestURI(str)
	if err != nil {
		return false
	}

	if (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" {
		return true
	}
	return false
}

func isContains(str string, rule Rule) bool {
	return strings.Contains(str, rule.Param)
}

func excludes(str string, rule Rule) bool {
	return !strings.Contains(str, rule.Param)
}

func isPrefix(str string, rule Rule) bool {
	return strings.HasPrefix(str, rule.Param)
}

func isSuffix(str string, rule Rule) bool {
	return strings.HasSuffix(str, rule.Param)
}

func isNumeric(str string, _ Rule) bool {
	_, err := strconv.ParseFloat(str, 64)
	return err == nil
}

func isUUID(str string, _ Rule) bool {
	_, err := uuid.Parse(str)
	return err == nil
}

var regNZUintRegexp = regexp.MustCompile(`^[1-9][0-9]*$`)

func notZeroUint(fc *FieldContext, rule Rule) error {
	v := fc.Value()
	if !v.IsValid() {
		return fc.UnsupportedType(rule)
	}

	switch v.Kind() {
	case reflect.String:
		s := v.String()
		if s == "" || regNZUintRegexp.MatchString(s) {
			return nil
		}
		return fc.Error(rule.Name, s)
	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		num := v.Int()
		if num > 0 {
			return nil
		}
		return fc.Error(rule.Name, num)
	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64:
		num := v.Uint()
		if num > 0 {
			return nil
		}
		return fc.Error(rule.Name, num)
	case reflect.Array, reflect.Slice:
		return internal.CheckSliceElem(fc, rule, notZeroUint)
	case reflect.Map:
		return internal.CheckMapValue(fc, rule, notZeroUint)
	default:
		return fc.UnsupportedType(rule)
	}
}

func validatePort(fc *FieldContext, rule Rule) error {
	v := fc.Value()
	if !v.IsValid() {
		return fc.UnsupportedType(rule)
	}

	switch v.Kind() {
	case reflect.String:
		s := v.String()
		if s == "" {
			return nil
		}
		num, err := strconv.ParseUint(s, 10, 64)
		if err == nil && num > 0 && num < 65535 {
			return nil
		}
		return fc.Error(rule.Name, internal.ShortErrStr(s))
	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		num := v.Int()
		if num > 0 && num < 65535 {
			return nil
		}
		return fc.Error(rule.Name, num)
	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64:
		num := v.Uint()
		if num > 0 && num < 65535 {
			return nil
		}
		return fc.Error(rule.Name, num)
	case reflect.Array, reflect.Slice:
		return internal.CheckSliceElem(fc, rule, validatePort)
	case reflect.Map:
		return internal.CheckMapValue(fc, rule, validatePort)
	default:
		return fc.UnsupportedType(rule)
	}
}

func validateIP(fc *FieldContext, rule Rule) error {
	v := fc.Value()
	if !v.IsValid() {
		return fc.UnsupportedType(rule)
	}

	switch v.Kind() {
	case reflect.String:
		s := v.String()
		if s == "" {
			return nil
		}
		ip := net.ParseIP(s)
		// 只判断格式正确
		if ip != nil {
			return nil
		}
		return fc.Error(rule.Name, internal.ShortErrStr(s))
	case reflect.Array, reflect.Slice:
		if v.Len() == 0 {
			return nil
		}
		if v.Type().Elem().Kind() == reflect.Uint8 {
			s := string(v.Bytes())
			ip := net.ParseIP(s)
			if ip != nil {
				return nil
			}
			return fc.Error(rule.Name, internal.ShortErrStr(s))
		}
		return internal.CheckSliceElem(fc, rule, validateIP)
	case reflect.Map:
		return internal.CheckMapValue(fc, rule, validateIP)
	default:
		return fc.UnsupportedType(rule)
	}
}

func validateIPV4(fc *FieldContext, rule Rule) error {
	v := fc.Value()
	if !v.IsValid() {
		return fc.UnsupportedType(rule)
	}

	switch v.Kind() {
	case reflect.String:
		s := v.String()
		if s == "" {
			return nil
		}
		ip := net.ParseIP(s)
		if ip != nil && ip.To4() != nil {
			return nil
		}
		return fc.Error(rule.Name, internal.ShortErrStr(s))
	case reflect.Array, reflect.Slice:
		if v.Len() == 0 {
			return nil
		}
		if v.Type().Elem().Kind() == reflect.Uint8 {
			s := string(v.Bytes())
			ip := net.ParseIP(s)
			if ip != nil && ip.To4() != nil {
				return nil
			}
			return fc.Error(rule.Name, internal.ShortErrStr(s))
		}
		return internal.CheckSliceElem(fc, rule, validateIPV4)
	case reflect.Map:
		return internal.CheckMapValue(fc, rule, validateIPV4)
	default:
		return fc.UnsupportedType(rule)
	}
}

func validateIPV6(fc *FieldContext, rule Rule) error {
	v := fc.Value()
	if !v.IsValid() {
		return fc.UnsupportedType(rule)
	}

	switch v.Kind() {
	case reflect.String:
		s := v.String()
		if s == "" {
			return nil
		}
		ip := net.ParseIP(s)
		if ip != nil && ip.To4() == nil {
			return nil
		}
		return fc.Error(rule.Name, internal.ShortErrStr(s))
	case reflect.Array, reflect.Slice:
		if v.Len() == 0 {
			return nil
		}
		if v.Type().Elem().Kind() == reflect.Uint8 {
			s := string(v.Bytes())
			ip := net.ParseIP(s)
			if ip != nil && ip.To4() == nil {
				return nil
			}
			return fc.Error(rule.Name, internal.ShortErrStr(s))
		}
		return internal.CheckSliceElem(fc, rule, validateIPV6)
	case reflect.Map:
		return internal.CheckMapValue(fc, rule, validateIPV6)
	default:
		return fc.UnsupportedType(rule)
	}
}

func isRegexp(str string, rule Rule) bool {
	value := zcache.Regexp.Get1(rule.Param)
	return value.Err == nil && value.Regexp.MatchString(str)
}
