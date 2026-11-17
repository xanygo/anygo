package internal

import (
	"fmt"
	"reflect"

	"github.com/xanygo/anygo/internal/zcache"
)

func ByRegexp(regText string) ValidationFunc {
	var checker ValidationFunc
	checker = func(fc *FieldContext, rule Rule) error {
		regexpValue := zcache.Regexp.Get1(regText)
		if regexpValue.Err != nil {
			return fc.InvalidParam(rule, regexpValue.Err)
		}
		reg := regexpValue.Regexp
		v := fc.Value()
		if !v.IsValid() {
			return fc.UnsupportedType(rule)
		}

		switch v.Kind() {
		case reflect.String:
			s := v.String()
			if s == "" || reg.MatchString(s) {
				return nil
			}
			return fc.Error(rule.Name, ShortErrStr(s))
		case reflect.Slice, reflect.Array:
			length := v.Len()
			if length == 0 {
				return nil
			}
			if v.Type().Elem().Kind() == reflect.Uint8 {
				bf := v.Bytes()
				if reg.Match(bf) {
					return nil
				}
				return fc.Error(rule.Name, ShortErrStr(string(bf)))
			}
			for i := 0; i < length; i++ {
				sv := v.Index(i)
				nc := fc.Child(sv)
				if err := checker(nc, rule); err != nil {
					return err
				}
			}
			return nil
		case reflect.Map:
			length := v.Len()
			if length == 0 {
				return nil
			}
			for _, key := range v.MapKeys() {
				sv := v.MapIndex(key)
				nc := fc.Child(sv)
				if err := checker(nc, rule); err != nil {
					return err
				}
			}
			return nil
		default:
			return fc.UnsupportedType(rule)
		}
	}
	return checker
}

func ByMatchString(match func(string, Rule) bool) ValidationFunc {
	var checker ValidationFunc
	checker = func(fc *FieldContext, rule Rule) error {
		v := fc.Value()
		if !v.IsValid() {
			return fc.UnsupportedType(rule)
		}

		switch v.Kind() {
		case reflect.String:
			s := v.String()
			if s == "" || match(s, rule) {
				return nil
			}
			return fc.Error(rule.Name, ShortErrStr(s))
		case reflect.Slice, reflect.Array:
			length := v.Len()
			if length == 0 {
				return nil
			}
			if v.Type().Elem().Kind() == reflect.Uint8 {
				str := string(v.Bytes())
				if match(str, rule) {
					return nil
				}
				return fc.Error(rule.Name, ShortErrStr(str))
			}
			for i := 0; i < length; i++ {
				sv := v.Index(i)
				nc := fc.Child(sv)
				if err := checker(nc, rule); err != nil {
					return err
				}
			}
			return nil
		case reflect.Map:
			return CheckMapValue(fc, rule, checker)
		default:
			return fc.UnsupportedType(rule)
		}
	}
	return checker
}

func CheckSliceElem(fc *FieldContext, rule Rule, checker ValidationFunc) error {
	v := fc.Value()
	length := v.Len()
	if length == 0 {
		return nil
	}
	for i := 0; i < length; i++ {
		sv := v.Index(i)
		nc := fc.Child(sv)
		if err := checker(nc, rule); err != nil {
			return err
		}
	}
	return nil
}

func CheckMapValue(fc *FieldContext, rule Rule, checker ValidationFunc) error {
	v := fc.Value()
	length := v.Len()
	if length == 0 {
		return nil
	}
	for _, key := range v.MapKeys() {
		sv := v.MapIndex(key)
		nc := fc.Child(sv)
		if err := checker(nc, rule); err != nil {
			return err
		}
	}
	return nil
}

func CheckInterfaceValue(fc *FieldContext, rule Rule, checker ValidationFunc) error {
	v := fc.Value()
	length := v.Len()
	if length == 0 {
		return nil
	}
	for _, key := range v.MapKeys() {
		sv := v.MapIndex(key)
		nc := fc.Child(sv)
		if err := checker(nc, rule); err != nil {
			return err
		}
	}
	return nil
}

func ShortErrStr(s string) string {
	length := len(s)
	if length < 16 {
		return fmt.Sprintf("%q", s)
	}
	if length < 33 {
		return fmt.Sprintf("%q (len=%d)", s, length)
	}
	return fmt.Sprintf("%q... (len=%d)", s[:31], length)
}
