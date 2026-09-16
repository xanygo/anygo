// Copyright(C) 2023 github.com/xanygo  All Rights Reserved.
// Author: hidu <duv123+git@gmail.com>
// Date: 2023/9/8

package xt

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/xanygo/anygo/cli/xcolor"
	"github.com/xanygo/anygo/internal/zreflect"
)

func Equal[T any](t Testing, actual T, expected T, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if equal(expected, actual) {
		return
	}
	var zero T
	txts := Labeleds{
		Labeled{Label: "Type", Content: fmt.Sprintf("%T", zero)},
	}
	txts = append(txts, doDiff(actual, expected)...)
	Fail(t, "Not equal", txts, msgAndArgs...)
}

func NotEqual[T any](t Testing, actual T, expected T, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if !equal(expected, actual) {
		return
	}
	var zero T
	txts := Labeleds{
		Labeled{Label: "Type", Content: fmt.Sprintf("%T", zero)},
		Labeled{Label: "Actual", Content: prettyGoValue(actual)},
	}
	Fail(t, "Should not equal", txts, msgAndArgs...)
}

func Slice[T any](v ...T) []T {
	return v
}

func Less[T cmp.Ordered](t Testing, x T, y T, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if cmp.Compare(x, y) != -1 {
		msg := fmt.Sprintf(`"%v" is not less than "%v"`, x, y)
		Fail(t, msg, nil, msgAndArgs...)
	}
}

func LessOrEqual[T cmp.Ordered](t Testing, x T, y T, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if cmp.Compare(x, y) == 1 {
		msg := fmt.Sprintf(`"%v" is not less than or equal to "%v"`, x, y)
		Fail(t, msg, nil, msgAndArgs...)
	}
}

func Greater[T cmp.Ordered](t Testing, x T, y T, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if cmp.Compare(x, y) != 1 {
		msg := fmt.Sprintf(`"%v" is not greater than "%v"`, x, y)
		Fail(t, msg, nil, msgAndArgs...)
	}
}

func GreaterOrEqual[T cmp.Ordered](t Testing, x T, y T, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if cmp.Compare(x, y) == -1 {
		msg := fmt.Sprintf(`"%v" is not greater than or equal to "%v"`, x, y)
		Fail(t, msg, nil, msgAndArgs...)
	}
}

func Error(t Testing, err error, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if err != nil {
		return
	}
	msg := "An error is expected but got nil"
	Fail(t, msg, nil, msgAndArgs...)
}

func NoError(t Testing, err error, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if err == nil {
		return
	}

	txts := Labeleds{
		Labeled{Label: "Actual", Content: zreflect.DumpString(err)},
	}

	msg := fmt.Sprintf("Received unexpected error: %s", errorText(err))
	Fail(t, msg, txts, msgAndArgs...)
}

func ErrorIs(t Testing, err error, target error, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if errors.Is(err, target) {
		return
	}

	txts := Labeleds{
		Labeled{Label: "Actual", Content: errorText(err)},
		Labeled{Label: "Target", Content: errorText(err)},
		Labeled{Label: "ActualDump", Content: zreflect.DumpString(err)},
		Labeled{Label: "TargetDump", Content: zreflect.DumpString(target)},
	}

	msg := fmt.Sprintf("errors.Is(%#v, %q) = false, want true", err, target)
	Fail(t, msg, txts, msgAndArgs...)
}

func ErrorNot(t Testing, err error, target error, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if !errors.Is(err, target) {
		return
	}
	txts := Labeleds{
		Labeled{Label: "Actual", Content: errorText(err)},
		Labeled{Label: "Target", Content: errorText(err)},
		Labeled{Label: "ActualDump", Content: zreflect.DumpString(err)},
		Labeled{Label: "TargetDump", Content: zreflect.DumpString(target)},
	}

	msg := fmt.Sprintf("errors.Is(%T, %T) = true, want false", err, target)
	Fail(t, msg, txts, msgAndArgs...)
}

func ErrorContains(t Testing, err error, substr string, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if err == nil {
		Fail(t, "error is nil", nil, msgAndArgs...)
		return
	}
	et := err.Error()
	if strings.Contains(et, substr) {
		return
	}
	msg := fmt.Sprintf("error %q should contains %q", et, substr)
	Fail(t, msg, nil, msgAndArgs...)
}

func ErrorNotContains(t Testing, err error, substr string, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	et := err.Error()
	if !strings.Contains(et, substr) {
		return
	}
	msg := fmt.Sprintf("error %q should not contains %q", et, substr)
	Fail(t, msg, nil, msgAndArgs...)
}

func True(t Testing, actual bool, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if actual {
		return
	}
	Fail(t, "Should be true", nil, msgAndArgs...)
}

func False(t Testing, actual bool, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if !actual {
		return
	}
	Fail(t, "Should be false", nil, msgAndArgs...)
}

func Nil(t Testing, actual any, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if isNil(actual) {
		return
	}
	txts := Labeleds{
		Labeled{Label: "Type", Content: fmt.Sprintf("%T", actual)},
		Labeled{Label: "Actual", Content: zreflect.DumpString(actual)},
	}
	Fail(t, "Expected nil, but not", txts, msgAndArgs...)
}

func NotNil(t Testing, actual any, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if !isNil(actual) {
		return
	}
	txts := Labeleds{
		Labeled{Label: "Type", Content: fmt.Sprintf("%T", actual)},
		Labeled{Label: "Actual", Content: fmt.Sprintf("%#v", actual)},
	}
	Fail(t, "Expected not nil value, but nil", txts, msgAndArgs...)
}

func Empty(t Testing, actual any, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}

	if isEmpty(actual) {
		return
	}
	txts := Labeleds{
		Labeled{Label: "Type", Content: fmt.Sprintf("%T", actual)},
		Labeled{Label: "Actual", Content: zreflect.DumpString(actual)},
	}
	Fail(t, "Should be empty, but not", txts, msgAndArgs...)
}

func NotEmpty(t Testing, actual any, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if !isEmpty(actual) {
		return
	}
	txts := Labeleds{
		Labeled{Label: "Type", Content: fmt.Sprintf("%T", actual)},
		Labeled{Label: "Actual", Content: zreflect.DumpString(actual)},
	}
	Fail(t, "Should NOT be empty, but not", txts, msgAndArgs...)
}

func HasPrefix[T StringByte](t Testing, actual T, prefix T, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if strings.HasPrefix(string(actual), string(prefix)) {
		return
	}
	txts := Labeleds{
		Labeled{Label: "Type", Content: fmt.Sprintf("%T", actual)},
		Labeled{Label: "Actual", Content: fmt.Sprintf("%q", actual)},
		Labeled{Label: "Prefix", Content: fmt.Sprintf("%q", prefix)},
	}
	Fail(t, "Should HasPrefix but not", txts, msgAndArgs...)
}

func NotPrefix[T StringByte](t Testing, actual T, prefix T, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if !strings.HasPrefix(string(actual), string(prefix)) {
		return
	}

	txts := Labeleds{
		Labeled{Label: "Type", Content: fmt.Sprintf("%T", actual)},
		Labeled{Label: "Actual", Content: fmt.Sprintf("%q", actual)},
		Labeled{Label: "Prefix", Content: fmt.Sprintf("%q", prefix)},
	}
	Fail(t, "Should not HasPrefix but yes", txts, msgAndArgs...)
}

func HasSuffix[T StringByte](t Testing, actual T, suffix T, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if strings.HasSuffix(string(actual), string(suffix)) {
		return
	}
	txts := Labeleds{
		Labeled{Label: "Type", Content: fmt.Sprintf("%T", actual)},
		Labeled{Label: "Actual", Content: fmt.Sprintf("%q", actual)},
		Labeled{Label: "Suffix", Content: fmt.Sprintf("%q", suffix)},
	}
	Fail(t, "Should HasSuffix but not", txts, msgAndArgs...)
}

func NotSuffix[T StringByte](t Testing, actual T, suffix T, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if !strings.HasSuffix(string(actual), string(suffix)) {
		return
	}
	txts := Labeleds{
		Labeled{Label: "Type", Content: fmt.Sprintf("%T", actual)},
		Labeled{Label: "Actual", Content: fmt.Sprintf("%q", actual)},
		Labeled{Label: "Suffix", Content: fmt.Sprintf("%q", suffix)},
	}
	Fail(t, "Should not HasSuffix but yes", txts, msgAndArgs...)
}

// Contains 是否包含子字符串
func Contains[T StringByte](t Testing, s T, substr T, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if strings.Contains(string(s), string(substr)) {
		return
	}
	txts := Labeleds{
		Labeled{Label: "Type", Content: fmt.Sprintf("%T", s)},
		Labeled{Label: "Actual", Content: fmt.Sprintf("%q", s)},
		Labeled{Label: "Substr", Content: fmt.Sprintf("%q", substr)},
	}
	Fail(t, "Should Contains but not", txts, msgAndArgs...)
}

// NotContains 是否不包含子字符串
func NotContains[T StringByte](t Testing, s T, substr T, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	index := strings.Index(string(s), string(substr))
	if index < 0 {
		return
	}

	txts := Labeleds{
		Labeled{Label: "Type", Content: fmt.Sprintf("%T", s)},
		Labeled{Label: "Actual", Content: fmt.Sprintf("%q", s)},
		Labeled{Label: "Index", Content: strconv.Itoa(index)},
		Labeled{Label: "Substr", Content: fmt.Sprintf("%q", substr)},
	}
	Fail(t, "Should not Contains but yes", txts, msgAndArgs...)
}

// AnyOf 获取到的值等于任意一个预期值
func AnyOf[T any](t Testing, actual T, collection []T, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	for _, item := range collection {
		if equal(actual, item) {
			return
		}
	}
	var zero T
	txts := Labeleds{
		Labeled{Label: "DataType", Content: fmt.Sprintf("%T", zero)},
		Labeled{Label: "Actual", Content: prettyGoValue(actual)},
		Labeled{Label: "Collection", Content: prettyGoValue(collection)},
	}
	Fail(t, fmt.Sprintf("Should AnyOf [%d]%T", len(collection), zero), txts, msgAndArgs...)
}

// NotAnyOf 获取到的值，必须不是任意输入值
func NotAnyOf[T any](t Testing, actual T, collection []T, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	var indexs []int
	for i, item := range collection {
		if equal(actual, item) {
			indexs = append(indexs, i)
		}
	}
	if len(indexs) == 0 {
		return
	}

	var zero T
	txts := Labeleds{
		Labeled{Label: "DataType", Content: fmt.Sprintf("%T", zero)},
		Labeled{Label: "Value", Content: prettyGoValue(actual)},
		Labeled{Label: "Index", Content: fmt.Sprintf("%v", indexs)},
		Labeled{Label: "Set", Content: prettyGoValue(collection)},
	}
	Fail(t, fmt.Sprintf("Should Not AnyOf [%d]%T", len(collection), zero), txts, msgAndArgs...)
}

// InSlice 检查 item 已在 collection 之中
func InSlice[S ~[]E, E comparable](t Testing, item E, collection S, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}

	if slices.Contains(collection, item) {
		return
	}

	txts := Labeleds{
		Labeled{Label: "Type", Content: fmt.Sprintf("%T", collection)},
		Labeled{Label: "Value", Content: fmt.Sprintf("%#v", item)},
		Labeled{Label: "Collection", Content: prettyGoValue(collection)},
	}
	Fail(t, "Should Contains but not", txts, msgAndArgs...)
}

// NotInSlice 判断 item 应该不在 collection 之中
func NotInSlice[S ~[]E, E comparable](t Testing, item E, collection S, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	index := slices.Index(collection, item)
	if index < 0 {
		return
	}

	txts := Labeleds{
		Labeled{Label: "Type", Content: fmt.Sprintf("%T", collection)},
		Labeled{Label: "Value", Content: fmt.Sprintf("%#v", item)},
		Labeled{Label: "Index", Content: strconv.Itoa(index)},
		Labeled{Label: "Collection", Content: prettyGoValue(collection)},
	}
	Fail(t, "Should NotContains but yes", txts, msgAndArgs...)
}

// AllInSlice 判断 items 所有元素都在 collection 之中
func AllInSlice[S ~[]E, E comparable](t Testing, items S, collection S, msgAndArgs ...any) {
	if len(items) == 0 {
		return
	}
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	mp := make(map[E]struct{}, len(collection))
	for _, v := range collection {
		mp[v] = struct{}{}
	}
	notIn := make(map[int]E)
	for index, item := range items {
		if _, has := mp[item]; !has {
			notIn[index] = item
		}
	}
	if len(notIn) == 0 {
		return
	}
	txts := Labeleds{
		Labeled{Label: "Type", Content: fmt.Sprintf("%T", collection)},
		Labeled{Label: "Items", Content: prettyGoValue(items)},
		Labeled{Label: "Collection", Content: prettyGoValue(collection)},
		Labeled{Label: "NotIn", Content: prettyGoValue(notIn)},
	}
	var zero E
	if len(notIn) == len(items) {
		Fail(t, fmt.Sprintf("All items are not in [%d]%T", len(items), zero), txts, msgAndArgs...)
	} else {
		Fail(t, fmt.Sprintf("%d items are not in [%d]%T", len(notIn), len(items), zero), txts, msgAndArgs...)
	}
}

// AllInSlice 判断 items 所有元素都不在 collection 之中
func AllNotInSlice[S ~[]E, E comparable](t Testing, items S, collection S, msgAndArgs ...any) {
	if len(items) == 0 {
		return
	}
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	mp := make(map[E]struct{}, len(collection))
	for _, v := range collection {
		mp[v] = struct{}{}
	}
	in := make(map[int]E)
	for index, item := range items {
		if _, has := mp[item]; has {
			in[index] = item
		}
	}
	if len(in) == 0 {
		return
	}
	txts := Labeleds{
		Labeled{Label: "Type", Content: fmt.Sprintf("%T", collection)},
		Labeled{Label: "Items", Content: prettyGoValue(items)},
		Labeled{Label: "Collection", Content: prettyGoValue(collection)},
		Labeled{Label: "Contains", Content: prettyGoValue(in)},
	}
	var zero E
	if len(in) == len(items) {
		Fail(t, fmt.Sprintf("All items are in [%d]%T", len(items), zero), txts, msgAndArgs...)
	} else {
		Fail(t, fmt.Sprintf("%d items are in [%d]%T", len(in), len(items), zero), txts, msgAndArgs...)
	}
}

// SortEqual 将两个 slice 排序后比较内容是否一样
func SortEqual[S ~[]E, E cmp.Ordered](t Testing, actual S, expected S, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	expected = slices.Clone(expected)
	slices.Sort(expected)

	actual = slices.Clone(actual)
	slices.Sort(actual)
	if equal(expected, actual) {
		return
	}

	txts := Labeleds{
		Labeled{Label: "Type", Content: fmt.Sprintf("%T", actual)},
	}
	txts = append(txts, doDiff(actual, expected)...)
	Fail(t, "Sorted Slice not equal", txts, msgAndArgs...)
}

// SortNotEqual 将两个 slice 排序后比较内容是否不一样
func SortNotEqual[S ~[]E, E cmp.Ordered](t Testing, actual S, expected S, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	expected = slices.Clone(expected)
	slices.Sort(expected)

	actualClone := slices.Clone(actual)
	slices.Sort(actualClone)
	if !equal(expected, actualClone) {
		return
	}

	txts := Labeleds{
		Labeled{Label: "Type", Content: fmt.Sprintf("%T", actualClone)},
		Labeled{Label: "Actual", Content: zreflect.DumpString(actualClone)},
	}
	Fail(t, "Sorted Slice should not equal", txts, msgAndArgs...)
}

func SamePtr(t Testing, actual any, expected any, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if samePointers(expected, actual) {
		return
	}

	txts := Labeleds{
		Labeled{Label: "Type", Content: fmt.Sprintf("%T", actual)},
		Labeled{Label: "Actual", Content: fmt.Sprintf("%p %#v", actual, actual)},
		Labeled{Label: "Expected", Content: fmt.Sprintf("%p %#v", expected, expected)},
	}
	Fail(t, "Not same Ptr", txts, msgAndArgs...)
}

func NotSamePtr(t Testing, expected any, actual any, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	if !samePointers(expected, actual) {
		return
	}

	txts := Labeleds{
		Labeled{Label: "Type", Content: fmt.Sprintf("%T", actual)},
		Labeled{Label: "Actual", Content: fmt.Sprintf("%p %#v", actual, actual)},
		Labeled{Label: "ExpectedNot", Content: fmt.Sprintf("%p %#v", expected, expected)},
	}
	Fail(t, "Sholud Not same Ptr", txts, msgAndArgs...)
}

func Len(t Testing, object any, length int, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}

	l, ok := getLen(object)
	if !ok {
		txts := Labeleds{
			Labeled{Label: "Type", Content: fmt.Sprintf("%T", object)},
			Labeled{Label: "Value", Content: prettyGoValue(object)},
			Labeled{Label: "Expected", Content: strconv.Itoa(length)},
		}
		Fail(t, "Could not be applied builtin len()", txts, msgAndArgs...)
		return
	}
	if l == length {
		return
	}
	txts := Labeleds{
		Labeled{Label: "Type", Content: fmt.Sprintf("%T", object)},
		Labeled{Label: "Actual", Content: strconv.Itoa(l)},
		Labeled{Label: "Expected", Content: strconv.Itoa(length)},
		Labeled{Label: "Value", Content: zreflect.DumpString(object)},
	}
	msg := fmt.Sprintf(`should have %d item(s), but has %d`, length, l)
	Fail(t, msg, txts, msgAndArgs...)
}

func Panic(t Testing, fn func(), msgAndArgs ...any) {
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
	Fail(t, "func should panic", nil, msgAndArgs...)
}

func NoPanic(t Testing, fn func(), msgAndArgs ...any) {
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
	if re == nil {
		return
	}
	msg := fmt.Sprintf("func panic: %v", re)
	Fail(t, msg, nil, msgAndArgs...)
}

func Fail(t Testing, failureMessage string, labels []Labeled, msgAndArgs ...any) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}

	var content Labeleds

	if n, ok := t.(interface{ Name() string }); ok {
		content = append(content, Labeled{Label: "Test", Content: n.Name()})
	}

	content = append(content, Labeled{Label: xcolor.RedString("Error"), Content: failureMessage})
	message := messageFromMsgAndArgs(msgAndArgs...)
	if len(message) > 0 {
		content = append(content, Labeled{Label: "Messages", Content: message})
	}
	content = append(content, labels...)

	t.Fatalf("\n%s\n", content.String())
}

func messageFromMsgAndArgs(msgAndArgs ...any) string {
	if len(msgAndArgs) == 0 || msgAndArgs == nil {
		return ""
	}
	if len(msgAndArgs) == 1 {
		msg := msgAndArgs[0]
		if msgAsStr, ok := msg.(string); ok {
			return msgAsStr
		}
		return fmt.Sprintf("%+v", msg)
	}
	if len(msgAndArgs) > 1 {
		return fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
	return ""
}
