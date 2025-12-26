//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-05

package resp3

import (
	"errors"
	"fmt"
	"strconv"
)

func ToInt(result Element, err error) (int, error) {
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

func ToString(result Element, err error) (string, error) {
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

func ToBytes(result Element, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	switch dv := result.(type) {
	case Null:
		return nil, nil
	case SimpleString:
		return dv.ToBytes(), nil
	case BulkString:
		return dv.ToBytes(), nil
	default:
		return nil, fmt.Errorf("%w: ToString %#v(%T)", ErrInvalidReply, dv, dv)
	}
}

func elementAsArray(e Element) ([]Element, error) {
	switch dv := e.(type) {
	case Array:
		return dv, nil
	case Set:
		return dv, nil
	case Push:
		return dv, nil
	default:
		return nil, fmt.Errorf("%w,not array reply: %T", ErrInvalidReply, e)
	}
}

func ToStringSlice(e Element, err error, expectLen int) ([]string, error) {
	if err != nil {
		return nil, err
	}
	arr, err := elementAsArray(e)
	if err != nil {
		return nil, err
	}
	if expectLen > 0 && len(arr) != expectLen {
		return nil, fmt.Errorf("array expect %d elements, got %d", expectLen, len(arr))
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

func ToPtrStringSlice(e Element, err error, expectLen int) ([]*string, error) {
	if err != nil {
		return nil, err
	}
	arr, err := elementAsArray(e)
	if err != nil {
		return nil, err
	}
	if expectLen > 0 && len(arr) != expectLen {
		return nil, fmt.Errorf("array expect %d elements, got %d", expectLen, len(arr))
	}
	list := make([]*string, 0, len(arr))
	for _, item := range arr {
		switch dv := item.(type) {
		case SimpleString:
			val := dv.String()
			list = append(list, &val)
		case BulkString:
			val := dv.String()
			list = append(list, &val)
		case Null:
			list = append(list, nil)
		default:
			return nil, fmt.Errorf("%w: ToStringSlice %#v(%T)", ErrInvalidReply, dv, dv)
		}
	}
	return list, nil
}

func ToBool(result Element, err error) (bool, error) {
	if err != nil {
		if errors.Is(err, ErrNil) {
			return false, nil
		}
		return false, err
	}
	switch dv := result.(type) {
	case Boolean:
		return dv.Bool(), nil
	default:
		return false, fmt.Errorf("%w: ToOkBool %#v(%T)", ErrInvalidReply, dv, dv)
	}
}

func ToOkBool(result Element, err error) (bool, error) {
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

func ToIntBool(result Element, err error, ok int) (bool, error) {
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

func ToIntBools(e Element, err error, expectLen int, ok int) ([]bool, error) {
	if err != nil {
		return nil, err
	}
	arr, err := elementAsArray(e)
	if err != nil {
		return nil, err
	}
	if expectLen > 0 && len(arr) != expectLen {
		return nil, fmt.Errorf("array expect %d elements, got %d", expectLen, len(arr))
	}
	result := make([]bool, expectLen)
	for i := 0; i < expectLen; i++ {
		result[i], err = ToIntBool(arr[i], nil, ok)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func ToOkStatus(result Element, err error) error {
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

func ToInt64(result Element, err error) (int64, error) {
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

func ToUint64(result Element, err error) (uint64, error) {
	if err != nil {
		return 0, err
	}
	switch dv := result.(type) {
	case Integer:
		return uint64(dv.Int64()), nil
	case BigNumber:
		return dv.BigInt().Uint64(), nil
	case SimpleString:
		return dv.ToUint64()
	case BulkString:
		return dv.ToUint64()
	default:
		return 0, fmt.Errorf("%w: ToUint64 %#v(%T)", ErrInvalidReply, result, result)
	}
}

func ToInt64Slice(e Element, err error, expectLen int) ([]int64, error) {
	if err != nil {
		return nil, err
	}
	arr, err := elementAsArray(e)
	if err != nil {
		return nil, err
	}
	if expectLen > 0 && len(arr) != expectLen {
		return nil, fmt.Errorf("array expect %d elements, got %d", expectLen, len(arr))
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
			return nil, fmt.Errorf("%w: ToInt64Slice_1 %#v(%T)", ErrInvalidReply, e, e)
		}
	}
	return list, nil
}

// ToPtrInt64Slice 解析出允许 null 值的 []*int64 结果
func ToPtrInt64Slice(e Element, err error) ([]*int64, error) {
	if err != nil {
		return nil, err
	}
	arr, err := elementAsArray(e)
	if err != nil {
		return nil, err
	}
	list := make([]*int64, 0, len(arr))
	for _, item := range arr {
		switch dv := item.(type) {
		case Integer:
			num := dv.Int64()
			list = append(list, &num)
		case BigNumber:
			num := dv.Int64()
			list = append(list, &num)
		case SimpleString:
			num, err1 := dv.ToInt64()
			if err1 != nil {
				return nil, err1
			}
			list = append(list, &num)
		case BulkString:
			num, err1 := dv.ToInt64()
			if err1 != nil {
				return nil, err1
			}
			list = append(list, &num)
		case Null:
			list = append(list, nil)
		default:
			return nil, fmt.Errorf("%w: ToInt64Slice_1 %#v(%T)", ErrInvalidReply, e, e)
		}
	}
	return list, nil
}

func ToFloat64(result Element, err error) (float64, error) {
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

func ToFloat64Slice(e Element, err error, expectLen int) ([]float64, error) {
	if err != nil {
		return nil, err
	}
	arr, err := elementAsArray(e)
	if err != nil {
		return nil, fmt.Errorf("%w: ToFloat64Slice_0", err)
	}
	if expectLen > 0 && len(arr) != expectLen {
		return nil, fmt.Errorf("array expect %d elements, got %d", expectLen, len(arr))
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

func ToStringMapWithKeys(e Element, err error, keys []string) (map[string]string, error) {
	if err != nil {
		return nil, err
	}
	arr, err := elementAsArray(e)
	if err != nil {
		return nil, fmt.Errorf("%w: ToStringMapWithKeys_0", err)
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

func asMap(e Element, err error) (map[Element]Element, error) {
	if err != nil {
		return nil, err
	}
	switch rv := e.(type) {
	case Map:
		return rv, nil
	case Attribute:
		return rv, nil
	default:
		return nil, fmt.Errorf("%w: asMap %#v", ErrInvalidReply, e)
	}
}

func ToStringMap(e Element, err error) (map[string]string, error) {
	mp, err := asMap(e, err)
	if err != nil {
		return nil, err
	}
	return mapToStringMap(mp)
}

func ToStringAnyMap(e Element, err error) (map[string]any, error) {
	mp, err := asMap(e, err)
	if err != nil {
		return nil, err
	}

	return mapToStringAnyMap(mp)
}

//	func ToMapFloat64(e Element, err error) (map[string]float64, error) {
//		if err != nil {
//			return nil, err
//		}
//		arr, err := elementAsArray(e)
//		if err != nil {
//			return nil, fmt.Errorf("%w: ToMapFloat64", err)
//		}
//		if len(arr)%2 != 0 {
//			return nil, fmt.Errorf("expected even number of keys, got %d", len(arr))
//		}
//		ret := make(map[string]float64, len(arr)/2)
//		for i := 0; i < len(arr); i += 2 {
//			member, err1 := ToString(arr[i], nil)
//			if err1 != nil {
//				return nil, err1
//			}
//			score, err2 := ToFloat64(arr[i+1], nil)
//			if err2 != nil {
//				return nil, err2
//			}
//			ret[member] = score
//		}
//		return ret, nil
//	}
func ToMapFloat64WithKeys(e Element, err error, keys []string) (map[string]float64, error) {
	if err != nil {
		return nil, err
	}
	arr, err := elementAsArray(e)
	if err != nil {
		return nil, fmt.Errorf("%w: ToMapFloat64WithKeys", err)
	}
	if len(arr) != len(keys) {
		return nil, fmt.Errorf("length not match, reply=%d, keys=%d", len(arr), len(keys))
	}
	ret := make(map[string]float64, len(arr))
	for i := 0; i < len(arr); i++ {
		if _, ok2 := arr[i].(Null); ok2 {
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

type maps interface {
	Map | Attribute | map[Element]Element
}

func mapToStringMap[T maps](m T) (map[string]string, error) {
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

func mapToStringAnyMap[T maps](m T) (map[string]any, error) {
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
		if err != nil && !IsRespError(err) {
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
		return vv, vv // error 类型
	case BulkString:
		return vv.String(), nil
	case BulkError:
		return vv, vv // // error 类型
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
		if err != nil && !IsRespError(err) {
			return nil, err
		}
		vs = append(vs, value)
	}
	return vs, nil
}

func ToAny(v Element, err error) (any, error) {
	if err != nil {
		return nil, err
	}
	switch vv := v.(type) {
	case Null:
		return nil, nil
	case SimpleString:
		return vv.String(), nil
	case SimpleError:
		return vv, vv // error 类型
	case BulkString:
		return vv.String(), nil
	case BulkError:
		return vv, vv // // error 类型
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
		return toAnySlice(vv)
	case Set:
		return toAnySlice(vv)
	case Push:
		return toAnySlice(vv)

	case Map:
		return toAnyMay(vv)
	case Attribute:
		return toAnyMay(vv)

	default:
		return nil, fmt.Errorf("unknown element %#v", vv)
	}
}

func ToAnySlice(e Element, err error) ([]any, error) {
	if err != nil {
		return nil, err
	}
	arr, err := elementAsArray(e)
	if err != nil {
		return nil, err
	}
	return toAnySlice(arr)
}

func toAnySlice(vv []Element) ([]any, error) {
	vs := make([]any, 0, len(vv))
	for _, e := range vv {
		value, err := ToAny(e, nil)
		if err != nil && !IsRespError(err) {
			return nil, err
		}
		vs = append(vs, value)
	}
	return vs, nil
}

func ToBoolSlice(e Element, err error, expectLen int) ([]bool, error) {
	if err != nil {
		return nil, err
	}
	arr, err := elementAsArray(e)
	if err != nil {
		return nil, err
	}
	if expectLen > 0 && len(arr) != expectLen {
		return nil, fmt.Errorf("expect %d elements, but got %d", expectLen, len(arr))
	}
	vs := make([]bool, 0, len(arr))
	for idx, item := range arr {
		switch tv := item.(type) {
		case Boolean:
			vs = append(vs, tv.Bool())
		default:
			return nil, fmt.Errorf("element[%d] not bool: %#v", idx, item)
		}
	}
	return vs, nil
}

func ToAnyMap(e Element, err error) (map[any]any, error) {
	if err != nil {
		return nil, err
	}
	switch m := e.(type) {
	case Map:
		return toAnyMay(m)
	case Attribute:
		return toAnyMay(m)
	default:
		return nil, fmt.Errorf("unknown element %#v", e)
	}
}

func toAnyMay(m map[Element]Element) (map[any]any, error) {
	if len(m) == 0 {
		return nil, nil
	}
	result := make(map[any]any, len(m))
	for k, v := range m {
		key, err1 := ToAny(k, nil)
		if err1 != nil && !IsRespError(err1) {
			return nil, err1
		}
		value, err2 := ToAny(v, nil)
		if err2 != nil && !IsRespError(err2) {
			return nil, err2
		}
		result[key] = value
	}
	return result, nil
}
