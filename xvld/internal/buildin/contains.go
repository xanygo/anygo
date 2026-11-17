package buildin

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/xanygo/anygo/xslice"
	"github.com/xanygo/anygo/xvld/internal"
)

func validateContains(fc *FieldContext, rule Rule) error {
	v := fc.Value()
	if !v.IsValid() {
		return fc.UnsupportedType(rule)
	}
	if rule.Param == "" {
		return fc.InvalidParam(rule, internal.ErrParamEmpty)
	}

	switch v.Kind() {
	case reflect.String:
		s := v.String()
		if s == "" {
			return nil
		}
		for _, item := range strings.Split(rule.Param, "|") {
			if strings.Contains(s, item) {
				return nil
			}
		}
		return fc.Error(rule.Name, rule.Param, internal.ShortErrStr(s))
	case reflect.Array, reflect.Slice:
		if v.Len() == 0 {
			return nil
		}
		return doSliceContains(fc, rule)
	default:
		return fc.UnsupportedType(rule)
	}
}

func doSliceContains(fc *FieldContext, rule Rule) error {
	v := fc.Value()
	switch v.Type().Elem().Kind() {
	case reflect.Uint8:
		s := string(v.Bytes())
		for _, item := range strings.Split(rule.Param, "|") {
			if strings.Contains(s, item) {
				return nil
			}
		}
		return fc.Error(rule.Name, rule.Param, internal.ShortErrStr(s))
	case reflect.String:
		mp := xslice.ToMap(strings.Split(rule.Param, "|"), true)
		head := make([]string, 0, 4)
		for i := 0; i < v.Len(); i++ {
			val := v.Index(i).String()
			if i < 3 {
				head = append(head, val)
			}
			if _, has := mp[val]; has {
				return nil
			}
		}
		return fc.Error(rule.Name, rule.Param, internal.ShortErrStr(strings.Join(head, ","))+"...")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		mp, err := parserRuleInt64Map(fc, rule)
		if err != nil {
			return err
		}
		head := make([]string, 0, 4)
		for i := 0; i < v.Len(); i++ {
			val := v.Index(i).Int()
			if i < 3 {
				head = append(head, strconv.FormatInt(val, 10))
			}
			if _, has := mp[val]; has {
				return nil
			}
		}
		return fc.Error(rule.Name, rule.Param, internal.ShortErrStr(strings.Join(head, ","))+"...")
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		mp, err := parserRuleUint64Map(fc, rule)
		if err != nil {
			return err
		}
		head := make([]string, 0, 4)
		for i := 0; i < v.Len(); i++ {
			val := v.Index(i).Uint()
			if i < 3 {
				head = append(head, strconv.FormatUint(val, 10))
			}
			if _, has := mp[val]; has {
				return nil
			}
		}
		return fc.Error(rule.Name, rule.Param, internal.ShortErrStr(strings.Join(head, ","))+"...")
	case reflect.Float32, reflect.Float64:
		mp, err := parserRuleFloat64Map(fc, rule)
		if err != nil {
			return err
		}
		head := make([]string, 0, 4)
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			val := item.Float()
			if i < 3 {
				head = append(head, strconv.FormatFloat(val, 'f', -1, item.Type().Bits()))
			}
			if _, has := mp[val]; has {
				return nil
			}
		}
		return fc.Error(rule.Name, rule.Param, internal.ShortErrStr(strings.Join(head, ","))+"...")
	case reflect.Bool:
		b, err := strconv.ParseBool(rule.Param)
		if err != nil {
			return fc.InvalidParam(rule, err)
		}
		head := make([]string, 0, 4)
		for i := 0; i < v.Len(); i++ {
			val := v.Index(i).Bool()
			if i < 3 {
				head = append(head, strconv.FormatBool(val))
			}
			if val == b {
				return nil
			}
		}
		return fc.Error(rule.Name, rule.Param, internal.ShortErrStr(strings.Join(head, ","))+"...")
	default:
		return fc.UnsupportedType(rule)
	}
}

func parserRuleInt64Map(fc *internal.FieldContext, rule Rule) (map[int64]struct{}, error) {
	arr := strings.Split(rule.Param, "|")
	data := make(map[int64]struct{}, len(arr))
	for _, item := range arr {
		num, err := strconv.ParseInt(item, 10, 64)
		if err != nil {
			return nil, fc.InvalidParam(rule, err)
		}
		data[num] = struct{}{}
	}
	return data, nil
}

func parserRuleUint64Map(fc *internal.FieldContext, rule Rule) (map[uint64]struct{}, error) {
	arr := strings.Split(rule.Param, "|")
	data := make(map[uint64]struct{}, len(arr))
	for _, item := range arr {
		num, err := strconv.ParseUint(item, 10, 64)
		if err != nil {
			return nil, fc.InvalidParam(rule, err)
		}
		data[num] = struct{}{}
	}
	return data, nil
}

func parserRuleFloat64Map(fc *internal.FieldContext, rule Rule) (map[float64]struct{}, error) {
	arr := strings.Split(rule.Param, "|")
	data := make(map[float64]struct{}, len(arr))
	for _, item := range arr {
		num, err := strconv.ParseFloat(item, 64)
		if err != nil {
			return nil, fc.InvalidParam(rule, err)
		}
		data[num] = struct{}{}
	}
	return data, nil
}

func validateExcludes(fc *FieldContext, rule Rule) error {
	v := fc.Value()
	if !v.IsValid() {
		return fc.UnsupportedType(rule)
	}
	if rule.Param == "" {
		return fc.InvalidParam(rule, internal.ErrParamEmpty)
	}

	switch v.Kind() {
	case reflect.String:
		s := v.String()
		if s == "" {
			return nil
		}
		for _, item := range strings.Split(rule.Param, "|") {
			if strings.Contains(s, item) {
				return fc.Error(rule.Name, item, internal.ShortErrStr(s))
			}
		}
		return nil
	case reflect.Array, reflect.Slice:
		if v.Len() == 0 {
			return nil
		}
		return doSliceExcludes(fc, rule)
	default:
		return fc.UnsupportedType(rule)
	}
}

func doSliceExcludes(fc *FieldContext, rule Rule) error {
	v := fc.Value()
	switch v.Type().Elem().Kind() {
	case reflect.Uint8:
		s := string(v.Bytes())
		for _, item := range strings.Split(rule.Param, "|") {
			if strings.Contains(s, item) {
				return fc.Error(rule.Name, item, internal.ShortErrStr(s))
			}
		}
		return nil
	case reflect.String:
		mp := xslice.ToMap(strings.Split(rule.Param, "|"), true)
		head := make([]string, 0, 4)
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i).String()
			if i < 3 {
				head = append(head, item)
			}
			if _, has := mp[item]; has {
				return fc.Error(rule.Name, item, internal.ShortErrStr(strings.Join(head, ","))+"...")
			}
		}
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		mp, err := parserRuleInt64Map(fc, rule)
		if err != nil {
			return err
		}
		head := make([]string, 0, 4)
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i).Int()
			if i < 3 {
				head = append(head, strconv.FormatInt(item, 10))
			}
			if _, has := mp[item]; has {
				return fc.Error(rule.Name, item, internal.ShortErrStr(strings.Join(head, ","))+"...")
			}
		}
		return nil
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		mp, err := parserRuleUint64Map(fc, rule)
		if err != nil {
			return err
		}
		head := make([]string, 0, 4)
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i).Uint()
			if i < 3 {
				head = append(head, strconv.FormatUint(item, 10))
			}
			if _, has := mp[item]; has {
				return fc.Error(rule.Name, item, internal.ShortErrStr(strings.Join(head, ","))+"...")
			}
		}
		return nil
	case reflect.Float32, reflect.Float64:
		mp, err := parserRuleFloat64Map(fc, rule)
		if err != nil {
			return err
		}
		head := make([]string, 0, 4)
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			val := item.Float()
			if i < 3 {
				head = append(head, strconv.FormatFloat(val, 'f', -1, item.Type().Bits()))
			}
			if _, has := mp[val]; has {
				return fc.Error(rule.Name, val, internal.ShortErrStr(strings.Join(head, ","))+"...")
			}
		}
		return nil
	case reflect.Bool:
		b, err := strconv.ParseBool(rule.Param)
		if err != nil {
			return fc.InvalidParam(rule, err)
		}
		head := make([]string, 0, 4)
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i).Bool()
			if i < 3 {
				head = append(head, strconv.FormatBool(item))
			}
			if item == b {
				return fc.Error(rule.Name, item, internal.ShortErrStr(strings.Join(head, ","))+"...")
			}
		}
		return nil
	default:
		return fc.UnsupportedType(rule)
	}
}
