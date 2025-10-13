//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-05

package resp3

import (
	"errors"
	"fmt"
	"strconv"
)

func ToInt(result Result, err error) (int, error) {
	if err != nil {
		return 0, err
	}
	switch dv := result.(type) {
	case Integer:
		return dv.Int(), nil
	case BigNumber:
		return int(dv.BigInt().Int64()), nil
	case Null:
		return 0, nil
	case SimpleString:
		return strconv.Atoi(dv.String())
	case BulkString:
		return strconv.Atoi(dv.String())
	default:
		return 0, fmt.Errorf("%w: ToInt %#v(%T)", ErrInvalidReply, dv, dv)
	}
}

func ToString(result Result, err error) (string, error) {
	if err != nil {
		return "", err
	}
	switch dv := result.(type) {
	case Null:
		return "", nil
	case SimpleString:
		return dv.String(), nil
	case BulkString:
		return dv.String(), nil
	default:
		return "", fmt.Errorf("%w: ToString %#v(%T)", ErrInvalidReply, dv, dv)
	}
}

func elementAsArray(result Result) ([]Element, error) {
	switch dv := result.(type) {
	case Array:
		return dv, nil
	case Set:
		return dv, nil
	default:
		return nil, fmt.Errorf("not array reply: %T", result)
	}
}

func ToStringSlice(result Result, err error) ([]string, error) {
	if err != nil {
		return nil, err
	}
	arr, err := elementAsArray(result)
	if err != nil {
		return nil, err
	}
	list := make([]string, 0, len(arr))
	for _, item := range arr {
		switch dv := item.(type) {
		case SimpleString:
			list = append(list, dv.String())
		case BulkString:
			list = append(list, dv.String())
		default:
			return nil, fmt.Errorf("%w: ToStringSlice %#v(%T)", ErrInvalidReply, dv, dv)
		}
	}
	return list, nil
}

func ToOkBool(result Result, err error) (bool, error) {
	if err != nil {
		if errors.Is(err, ErrNil) {
			return false, nil
		}
		return false, err
	}
	switch dv := result.(type) {
	case SimpleString:
		return dv.String() == "OK", nil
	case BulkString:
		return dv.String() == "OK", nil
	default:
		return false, fmt.Errorf("%w: ToOkBool %#v(%T)", ErrInvalidReply, dv, dv)
	}
}

func ToIntBool(result Result, err error, ok int) (bool, error) {
	num, err1 := ToInt(result, err)
	if err1 != nil {
		return false, err1
	}
	switch num {
	case ok:
		return true, nil
	case 0:
		return false, nil
	default:
		return false, fmt.Errorf("%w: ToIntBool %#v", ErrInvalidReply, result)
	}
}

func ToOkStatus(result Result, err error) error {
	if err != nil {
		return err
	}
	switch dv := result.(type) {
	case SimpleString:
		if dv.String() == "OK" {
			return nil
		}
	case BulkString:
		if dv.String() == "OK" {
			return nil
		}
	default:
	}
	return fmt.Errorf("%w: ToOkStatus %#v(%T)", ErrInvalidReply, result, result)
}

func ToInt64(result Result, err error) (int64, error) {
	if err != nil {
		return 0, err
	}
	switch dv := result.(type) {
	case Integer:
		return dv.Int64(), nil
	case BigNumber:
		return dv.BigInt().Int64(), nil
	case SimpleString:
		return dv.ToInt64()
	case BulkString:
		return dv.ToInt64()
	default:
		return 0, fmt.Errorf("%w: ToInt64 %#v(%T)", ErrInvalidReply, result, result)
	}
}

func ToInt64Slice(result Result, err error) ([]int64, error) {
	if err != nil {
		return nil, err
	}
	arr, ok := result.(Array)
	if !ok {
		return nil, fmt.Errorf("%w: ToInt64Slice_0 %#v(%T)", ErrInvalidReply, result, result)
	}
	list := make([]int64, 0, len(arr))
	for _, item := range arr {
		switch dv := item.(type) {
		case Integer:
			list = append(list, dv.Int64())
		case BigNumber:
			list = append(list, dv.Int64())
		case SimpleString:
			num, err1 := dv.ToInt64()
			if err1 != nil {
				return nil, err1
			}
			list = append(list, num)
		case BulkString:
			num, err1 := dv.ToInt64()
			if err1 != nil {
				return nil, err1
			}
			list = append(list, num)
		default:
			return nil, fmt.Errorf("%w: ToInt64Slice_1 %#v(%T)", ErrInvalidReply, result, result)
		}
	}
	return list, nil
}

func ToFloat64(result Result, err error) (float64, error) {
	if err != nil {
		return 0, err
	}
	switch dv := result.(type) {
	case SimpleString:
		return dv.ToFloat64()
	case BulkString:
		return dv.ToFloat64()
	case Double:
		return dv.Float64(), nil
	default:
		return 0, fmt.Errorf("%w: ToFloat64 %#v(%T)", ErrInvalidReply, result, result)
	}
}

func ToFloat64Slice(result Result, err error) ([]float64, error) {
	if err != nil {
		return nil, err
	}
	arr, ok := result.(Array)
	if !ok {
		return nil, fmt.Errorf("%w: ToFloat64Slice_0 %#v(%T)", ErrInvalidReply, result, result)
	}
	list := make([]float64, 0, len(arr))
	for _, item := range arr {
		switch dv := item.(type) {
		case Double:
			list = append(list, dv.Float64())
		default:
			return nil, fmt.Errorf("%w: ToFloat64Slice_1 %#v(%T)", ErrInvalidReply, item, item)
		}
	}
	return list, nil
}

func ToStringMapWithKeys(result Result, err error, keys []string) (map[string]string, error) {
	if err != nil {
		return nil, err
	}
	arr, ok := result.(Array)
	if !ok {
		return nil, fmt.Errorf("%w: ToStringMapWithKeys_0 %#v(%T)", ErrInvalidReply, result, result)
	}
	if len(keys) != len(arr) {
		return nil, fmt.Errorf("expected %d keys, got %d", len(keys), len(arr))
	}
	m := make(map[string]string, len(keys))
	for idx, key := range keys {
		item := arr[idx]
		switch dv := item.(type) {
		case BulkString:
			m[key] = dv.String()
		case Null:
			continue
		default:
			return nil, fmt.Errorf("%w: ToStringMapWithKeys_1 %#v(%T)", ErrInvalidReply, key, key)
		}
	}
	return m, nil
}

func ToStringMap(result Result, err error) (map[string]string, error) {
	if err != nil {
		return nil, err
	}
	switch rv := result.(type) {
	case Map:
		return rv.ToStringMap()
	}
	return nil, nil
}

func ToMapFloat64(result Result, err error) (map[string]float64, error) {
	if err != nil {
		return nil, err
	}
	arr, ok := result.(Array)
	if !ok {
		return nil, fmt.Errorf("%w: ToMapFloat64 %#v(%T)", ErrInvalidReply, result, result)
	}
	if len(arr)%2 != 0 {
		return nil, fmt.Errorf("expected even number of keys, got %d", len(arr))
	}
	ret := make(map[string]float64, len(arr)/2)
	for i := 0; i < len(arr); i += 2 {
		member, err1 := ToString(arr[i], nil)
		if err1 != nil {
			return nil, err1
		}
		score, err2 := ToFloat64(arr[i+1], nil)
		if err2 != nil {
			return nil, err2
		}
		ret[member] = score
	}
	return ret, nil
}

func ToMapFloat64WithKeys(result Result, err error, keys []string) (map[string]float64, error) {
	if err != nil {
		return nil, err
	}
	arr, ok := result.(Array)
	if !ok {
		return nil, fmt.Errorf("%w: ToMapFloat64WithKeys %#v(%T)", ErrInvalidReply, result, result)
	}
	if len(arr) != len(keys) {
		return nil, fmt.Errorf("length not match, reply=%d, keys=%d", len(arr), len(keys))
	}
	ret := make(map[string]float64, len(arr))
	for i := 0; i < len(arr); i++ {
		if _, ok := arr[i].(Null); ok {
			continue
		}
		member := keys[i]
		score, err2 := ToFloat64(arr[i], nil)
		if err2 != nil {
			return nil, err2
		}
		ret[member] = score
	}
	return ret, nil
}

func mapToStringMap[T Map | Attribute](m T) (map[string]string, error) {
	if len(m) == 0 {
		return nil, nil
	}
	result := make(map[string]string, len(m))
	for k, v := range m {
		ks, ok1 := k.(fmt.Stringer)
		vs, ok2 := v.(fmt.Stringer)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("map: not string k-v %#v: %#v", k, v)
		}
		result[ks.String()] = vs.String()
	}

	return result, nil
}

func mapToStringAnyMap[T Map | Attribute](m T) (map[string]any, error) {
	return (stringAnyMapConverter{}).ToMap(m)
}

type stringAnyMapConverter struct{}

func (sc stringAnyMapConverter) ToMap(m map[Element]Element) (map[string]any, error) {
	if len(m) == 0 {
		return nil, nil
	}
	result := make(map[string]any, len(m))
	for k, v := range m {
		ks, ok1 := k.(fmt.Stringer)
		if !ok1 {
			return nil, fmt.Errorf("map: not string k-v %#v: %#v", k, v)
		}
		vs, err := sc.toAny(v)
		if err != nil {
			return nil, err
		}
		result[ks.String()] = vs
	}
	return result, nil
}

func (sc stringAnyMapConverter) toAny(v Element) (any, error) {
	switch vv := v.(type) {
	case Null:
		return nil, nil
	case SimpleString:
		return vv.String(), nil
	case SimpleError:
		return vv, nil // error 类型
	case BulkString:
		return vv.String(), nil
	case BulkError:
		return vv, nil // // error 类型
	case Double:
		return vv.Float64(), nil
	case Integer:
		return vv.Int(), nil
	case Boolean:
		return vv.Bool(), nil
	case BigNumber:
		return vv.Int64(), nil

	case VerbatimString:
		return vv, nil

	case Array:
		return sc.toAnySlice(vv)
	case Set:
		return sc.toAnySlice(vv)
	case Push:
		return sc.toAnySlice(vv)

	case Map:
		return sc.ToMap(vv)
	case Attribute:
		return sc.ToMap(vv)
	default:
		return nil, fmt.Errorf("unknown element %#v", vv)
	}
}

func (sc stringAnyMapConverter) toAnySlice(vv []Element) ([]any, error) {
	vs := make([]any, 0, len(vv))
	for _, e := range vv {
		value, err := sc.toAny(e)
		if err != nil {
			return nil, err
		}
		vs = append(vs, value)
	}
	return vs, nil
}

func ToAny(v Element) (any, error) {
	switch vv := v.(type) {
	case Null:
		return nil, nil
	case SimpleString:
		return vv.String(), nil
	case SimpleError:
		return vv, nil // error 类型
	case BulkString:
		return vv.String(), nil
	case BulkError:
		return vv, nil // // error 类型
	case Double:
		return vv.Float64(), nil
	case Integer:
		return vv.Int64(), nil
	case Boolean:
		return vv.Bool(), nil
	case BigNumber:
		return vv.Int64(), nil

	case VerbatimString:
		return vv, nil

	case Array:
		return ToAnySlice(vv)
	case Set:
		return ToAnySlice(vv)
	case Push:
		return ToAnySlice(vv)

	case Map:
		return ToAnyMap(vv)
	case Attribute:
		return ToAnyMap(vv)

	default:
		return nil, fmt.Errorf("unknown element %#v", vv)
	}
}

func ToAnySlice(vv []Element) ([]any, error) {
	vs := make([]any, 0, len(vv))
	for _, e := range vv {
		value, err := ToAny(e)
		if err != nil {
			return nil, err
		}
		vs = append(vs, value)
	}
	return vs, nil
}

func ToAnyMap(m map[Element]Element) (map[any]any, error) {
	if len(m) == 0 {
		return nil, nil
	}
	result := make(map[any]any, len(m))
	for k, v := range m {
		ks, err1 := ToAny(k)
		if err1 != nil {
			return nil, err1
		}
		vs, err2 := ToAny(v)
		if err2 != nil {
			return nil, err2
		}
		result[ks] = vs
	}
	return result, nil
}
