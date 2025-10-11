//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-01

package xredis

import (
	"fmt"

	"github.com/xanygo/anygo/store/xredis/resp3"
)

type Response struct {
	result Result
	err    error
}

func (rs Response) Result() (any, error) {
	if rs.err != nil {
		return nil, rs.err
	}
	return resp3.ToAny(rs.result)
}

func (rs Response) Err() error {
	return rs.err
}

func (rs Response) Int() (int, error) {
	return resp3.ToInt(rs.result, rs.err)
}

func (rs Response) Int64() (int64, error) {
	return resp3.ToInt64(rs.result, rs.err)
}

func (rs Response) Float64() (float64, error) {
	return resp3.ToFloat64(rs.result, rs.err)
}

func (rs Response) Float64Slice() ([]float64, error) {
	return resp3.ToFloat64Slice(rs.result, rs.err)
}

func (rs Response) String() (string, error) {
	return resp3.ToString(rs.result, rs.err)
}

func (rs Response) OKStatus() error {
	return resp3.ToOkStatus(rs.result, rs.err)
}

func (rs Response) StringSlice() ([]string, error) {
	return resp3.ToStringSlice(rs.result, rs.err)
}

func (rs Response) StringMap() (map[string]string, error) {
	return resp3.ToStringMap(rs.result, rs.err)
}

func (rs Response) StringAnyMap() (map[string]any, error) {
	return convert[map[string]any](rs.result, rs.err)
}

func (rs Response) Map() (map[any]any, error) {
	return convert[map[any]any](rs.result, rs.err)
}

func convert[T any](result Result, err error) (v T, e error) {
	if err != nil {
		return v, err
	}
	obj, err := resp3.ToAny(result)
	if err != nil {
		return v, err
	}
	mp, ok := obj.(T)
	if ok {
		return mp, nil
	}
	return v, fmt.Errorf("got type %T, expected %T", result, v)
}
