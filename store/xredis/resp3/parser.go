//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-01

package resp3

import (
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
		return 0, fmt.Errorf("unexpected response type: %T", dv)
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
		return "", fmt.Errorf("unexpected response type: %T", dv)
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
			return nil, fmt.Errorf("unexpected response type: %T", dv)
		}
	}
	return list, nil
}

func ToOkBool(result Result, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	switch dv := result.(type) {
	case SimpleString:
		return dv.String() == "OK", nil
	case BulkString:
		return dv.String() == "OK", nil
	default:
		return false, fmt.Errorf("unexpected response type: %T", dv)
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
		return false, fmt.Errorf("unexpected reply: %v", num)
	}
}

func ToOkError(result Result, err error) error {
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
	return fmt.Errorf("unexpected response type: %T", result)
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
		return 0, fmt.Errorf("unexpected response value: %v(%T)", dv, dv)
	}
}

func ToInt64Slice(result Result, err error) ([]int64, error) {
	if err != nil {
		return nil, err
	}
	arr, ok := result.(Array)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", result)
	}
	list := make([]int64, 0, len(arr))
	for _, item := range arr {
		switch dv := item.(type) {
		case Integer:
			list = append(list, dv.Int64())
		case BigNumber:
			list = append(list, dv.Int64())
		case SimpleString:
			num, err := dv.ToInt64()
			if err != nil {
				return nil, err
			}
			list = append(list, num)
		case BulkString:
			num, err := dv.ToInt64()
			if err != nil {
				return nil, err
			}
			list = append(list, num)
		default:
			return nil, fmt.Errorf("unexpected response type: %T", dv)
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
	default:
		return 0, fmt.Errorf("unexpected response type: %T", dv)
	}
}

func ToMapWithKeys(result Result, err error, keys []string) (map[string]string, error) {
	if err != nil {
		return nil, err
	}
	arr, ok := result.(Array)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", result)
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
		default:
			return nil, fmt.Errorf("unexpected response type: %T", dv)
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
