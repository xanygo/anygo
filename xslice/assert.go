//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-12-31

package xslice

import (
	"errors"
	"fmt"
	"slices"
)

func CheckLenIn[S ~[]T, T any](s S, err error, lengths ...int) error {
	if err != nil {
		return err
	}
	if len(lengths) == 0 {
		return errors.New("expect lengths must not be empty")
	}
	l := len(s)
	if slices.Contains(lengths, l) {
		return nil
	}
	if len(lengths) == 1 {
		return fmt.Errorf("expected length %d but got %d", lengths[0], l)
	}
	return fmt.Errorf("expected length one of %v but got %d", lengths, l)
}

func CheckLenBetween[S ~[]T, T any](s S, err error, min, max int) error {
	if err != nil {
		return err
	}
	l := len(s)
	if l < min || l > max {
		return fmt.Errorf("expected length between [%d,%d] got %d", min, max, l)
	}
	return nil
}

func CheckLenAtMost[S ~[]T, T any](s S, err error, max int) error {
	if err != nil {
		return err
	}

	l := len(s)
	if l <= max {
		return nil
	}
	return fmt.Errorf("length must be <= %d, got %d", max, l)
}

func CheckLenAtLeast[S ~[]T, T any](s S, err error, min int) error {
	if err != nil {
		return err
	}

	l := len(s)
	if l >= min {
		return nil
	}
	return fmt.Errorf("length must be >= %d, got %d", min, l)
}
