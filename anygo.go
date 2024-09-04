//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-25

package anygo

// Ternary 三元表达式
func Ternary[T any](cond bool, trueValue T, falseValue T) T {
	if cond {
		return trueValue
	}
	return falseValue
}

// Must 若 values 中任意一项是 error 并且 err != nil ，则 panic
func Must(values ...any) {
	for _, v := range values {
		if err, ok := v.(error); ok && err != nil {
			panic(err)
		}
	}
}

type Then struct {
	err error
}

func (t *Then) Then(fn func() error) *Then {
	if t.err == nil {
		t.err = fn()
	}
	return t
}

func (t *Then) Err() error {
	return t.err
}

func DoThen(fn func() error) *Then {
	return &Then{
		err: fn(),
	}
}
