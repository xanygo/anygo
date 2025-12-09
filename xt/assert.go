// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/9/8

package xt

import (
	"bytes"
	"cmp"
	"errors"
	"reflect"
	"slices"
	"strings"

	"github.com/xanygo/anygo/internal/zreflect"
)

func Equal[T any](t Testing, expected T, actual T) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if !equal(expected, actual) {
		t.Fatalf("Not equal: \n%s", sprintfDiff(expected, actual))
	}
}

func NotEqual[T any](t Testing, expected T, actual T) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if equal(expected, actual) {
		str := zreflect.DumpString(actual)
		t.Fatalf("Should not equal: %s", str)
	}
}

func Less[T cmp.Ordered](t Testing, x T, y T) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if cmp.Compare(x, y) != -1 {
		t.Fatalf(`"%v" is not less than "%v"`, x, y)
	}
}

func LessOrEqual[T cmp.Ordered](t Testing, x T, y T) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if cmp.Compare(x, y) == 1 {
		t.Fatalf(`"%v" is not less than or equal to "%v"`, x, y)
	}
}

func Greater[T cmp.Ordered](t Testing, x T, y T) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if cmp.Compare(x, y) != 1 {
		t.Fatalf(`"%v" is not greater than "%v"`, x, y)
	}
}

func GreaterOrEqual[T cmp.Ordered](t Testing, x T, y T) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if cmp.Compare(x, y) == -1 {
		t.Fatalf(`"%v" is not greater than or equal to "%v"`, x, y)
	}
}

func Error(t Testing, err error) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if err == nil {
		t.Fatalf("An error is expected but got nil.")
	}
}

func NoError(t Testing, err error) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if err != nil {
		t.Fatalf("Received unexpected error: %s", errorText(err))
	}
}

func ErrorIs(t Testing, err error, target error) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if errors.Is(err, target) {
		return
	}

	var expectedText string
	if target != nil {
		expectedText = target.Error()
	}

	chain := buildErrorChainString(err)
	t.Fatalf("Target error should be in err chain:\n"+
		"expected: %q\n"+
		"in chain: %s", expectedText, chain,
	)
}

func ErrorContains(t Testing, err error, substr string) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	et := err.Error()
	if strings.Contains(et, substr) {
		return
	}
	t.Fatalf("error %q should contains %q", et, substr)
}

func ErrorNotContains(t Testing, err error, substr string) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	et := err.Error()
	if !strings.Contains(et, substr) {
		return
	}
	t.Fatalf("error %q should not contains %q", et, substr)
}

func NotErrorIs(t Testing, err error, target error) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if !errors.Is(err, target) {
		return
	}

	var expectedText string
	if target != nil {
		expectedText = target.Error()
	}

	chain := buildErrorChainString(err)
	t.Fatalf("Target error should not be in err chain:\n"+
		"expected: %q\n"+
		"in chain: %s", expectedText, chain,
	)
}

func True(t Testing, got bool) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if !got {
		t.Fatalf("Should be true")
	}
}

func False(t Testing, got bool) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if got {
		t.Fatalf("Should be false")
	}
}

func Nil(t Testing, got any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if !isNil(got) {
		t.Fatalf("Expected nil, but got: %#v", got)
	}
}

func NotNil(t Testing, got any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if isNil(got) {
		t.Fatalf("Expected value not to be nil")
	}
}

func Empty(t Testing, got any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}

	if !isEmpty(got) {
		t.Fatalf("Should be empty, but was %v", got)
	}
}

func NotEmpty(t Testing, got any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if isEmpty(got) {
		t.Fatalf("Should NOT be empty, but was %v", got)
	}
}

func HasPrefix[T StringByte](t Testing, s T, prefix T) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	vt := reflect.ValueOf(s)
	switch vt.Kind() {
	case reflect.String:
		if strings.HasPrefix(vt.String(), reflect.ValueOf(prefix).String()) {
			return
		}
	case reflect.Slice:
		if bytes.HasPrefix(vt.Bytes(), reflect.ValueOf(prefix).Bytes()) {
			return
		}
	}
	t.Fatalf("Should HasPrefix but not\n content : %q\n prefix  : %q", s, prefix)
}

func NotPrefix[T StringByte](t Testing, s T, prefix T) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	vt := reflect.ValueOf(s)
	switch vt.Kind() {
	case reflect.String:
		if !strings.HasPrefix(vt.String(), reflect.ValueOf(prefix).String()) {
			return
		}
	case reflect.Slice:
		if !bytes.HasPrefix(vt.Bytes(), reflect.ValueOf(prefix).Bytes()) {
			return
		}
	}
	t.Fatalf("Should not HasPrefix but yes\n content : %q\n prefix  : %q", s, prefix)
}

func HasSuffix[T StringByte](t Testing, s T, prefix T) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	vt := reflect.ValueOf(s)
	switch vt.Kind() {
	case reflect.String:
		if strings.HasSuffix(vt.String(), reflect.ValueOf(prefix).String()) {
			return
		}
	case reflect.Slice:
		if bytes.HasSuffix(vt.Bytes(), reflect.ValueOf(prefix).Bytes()) {
			return
		}
	}
	t.Fatalf("Should HasSuffix but not\n content : %q\n prefix  : %q", s, prefix)
}

func NotSuffix[T StringByte](t Testing, s T, prefix T) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	vt := reflect.ValueOf(s)
	switch vt.Kind() {
	case reflect.String:
		if !strings.HasSuffix(vt.String(), reflect.ValueOf(prefix).String()) {
			return
		}
	case reflect.Slice:
		if !bytes.HasSuffix(vt.Bytes(), reflect.ValueOf(prefix).Bytes()) {
			return
		}
	}
	t.Fatalf("Should not HasSuffix but yes\n content : %q\n prefix  : %q", s, prefix)
}

func Contains[T StringByte](t Testing, s T, substr T) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	vt := reflect.ValueOf(s)
	switch vt.Kind() {
	case reflect.String:
		if strings.Contains(vt.String(), reflect.ValueOf(substr).String()) {
			return
		}
	case reflect.Slice:
		if bytes.Contains(vt.Bytes(), reflect.ValueOf(substr).Bytes()) {
			return
		}
	}
	t.Fatalf("%#v should not substr %#v", s, substr)
}

func NotContains[T StringByte](t Testing, s T, substr T) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	vt := reflect.ValueOf(s)
	switch vt.Kind() {
	case reflect.String:
		if !strings.Contains(vt.String(), reflect.ValueOf(substr).String()) {
			return
		}
	case reflect.Slice:
		if !bytes.Contains(vt.Bytes(), reflect.ValueOf(substr).Bytes()) {
			return
		}
	}
	t.Fatalf("%#v should not substr %#v", s, substr)
}

func SliceContains[S ~[]E, E comparable](t Testing, values S, item E) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if !slices.Contains(values, item) {
		t.Fatalf("%#v does not contains %#v", values, item)
	}
}

func SliceNotContains[S ~[]E, E comparable](t Testing, values S, item E) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if slices.Contains(values, item) {
		t.Fatalf("%#v should not contains %#v", values, item)
	}
}

// SliceSortEqual 将两个 slice 排序后比较内容是否一样
func SliceSortEqual[S ~[]E, E cmp.Ordered](t Testing, expected S, actual S) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	expected = slices.Clone(expected)
	slices.Sort(expected)

	actual = slices.Clone(actual)
	slices.Sort(actual)
	if !equal(expected, actual) {
		t.Fatalf("Not equal: \n%s", sprintfDiff(expected, actual))
	}
}

// SliceSortNotEqual 将两个 slice 排序后比较内容是否不一样
func SliceSortNotEqual[S ~[]E, E cmp.Ordered](t Testing, expected S, actual S) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	expected = slices.Clone(expected)
	slices.Sort(expected)

	actual = slices.Clone(actual)
	slices.Sort(actual)
	if equal(expected, actual) {
		str := zreflect.DumpString(actual)
		t.Fatalf("Values should not be equal:\n %s", str)
	}
}

func SamePtr(t Testing, expected any, actual any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if !samePointers(expected, actual) {
		t.Fatalf("Not same: \n"+
			"expected: %p %#v\n"+
			"actual  : %p %#v", expected, expected, actual, actual)
	}
}

func NotSamePtr(t Testing, expected any, actual any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if samePointers(expected, actual) {
		t.Fatalf("Expected and actual point to the same object: %p %#v", expected, expected)
	}
}

func Len(t Testing, object any, length int) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}

	l, ok := getLen(object)
	if !ok {
		t.Fatalf(`"%v" could not be applied builtin len()`, object)
		return
	}

	if l != length {
		t.Fatalf(`"%v" should have %d item(s), but has %d`, object, length, l)
	}
}

func Panic(t Testing, fn func()) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	var re any
	func() {
		defer func() {
			re = recover()
		}()
		fn()
	}()
	if re != nil {
		return
	}
	t.Fatalf("func %#v should panic", fn)
}
