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

func Must1[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func Must2[A any, B any](v1 A, v2 B, err error) (A, B) {
	if err != nil {
		panic(err)
	}
	return v1, v2
}

func Must3[A any, B any, C any](v1 A, v2 B, v3 C, err error) (A, B, C) {
	if err != nil {
		panic(err)
	}
	return v1, v2, v3
}

type Then struct {
	err error
}

func (t *Then) Then(fns ...func() error) *Then {
	for _, fn := range fns {
		if t.err == nil {
			t.err = fn()
		} else {
			break
		}
	}
	return t
}

func (t *Then) Err() error {
	return t.err
}

func DoThen(fns ...func() error) *Then {
	return (&Then{}).Then(fns...)
}
