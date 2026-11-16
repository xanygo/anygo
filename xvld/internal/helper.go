package internal

import (
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
			return fc.Error(rule.Name, s)
		case reflect.Slice, reflect.Array:
			length := v.Len()
			if length == 0 {
				return nil
			}
			if v.Elem().Kind() == reflect.Uint8 && reg.Match(v.Bytes()) {
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
			return fc.Error(rule.Name, s)
		case reflect.Slice, reflect.Array:
			length := v.Len()
			if length == 0 {
				return nil
			}
			if v.Elem().Kind() == reflect.Uint8 && match(string(v.Bytes()), rule) {
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
