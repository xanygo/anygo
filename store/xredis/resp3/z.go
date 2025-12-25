//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-05

package resp3

import "fmt"

type Z struct {
	Score  float64
	Member string
}

func ToZSlice(ret Element, err error) ([]Z, error) {
	if err != nil {
		return nil, err
	}
	arr, ok := ret.(Array)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", ret)
	}
	return arr.ToZSlice()
}

func ToZSliceFlat(ret Element, err error) ([]Z, error) {
	if err != nil {
		return nil, err
	}
	arr, ok := ret.(Array)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", ret)
	}
	return arr.ToZSliceFlat()
}
