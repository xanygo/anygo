//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-15

package xredis

import (
	"fmt"

	"github.com/xanygo/anygo/store/xredis/resp3"
)

type Cmder interface {
	Args() []any
	Name() string
	CmdResult
}

type CmdResult interface {
	SetReply(result resp3.Element, err error)
	SetErr(err error)
	ValueErr() (any, error)
	Err() error
	Value() any
}

type baseCmd struct {
	args []any
}

func (b *baseCmd) Name() string {
	if len(b.args) == 0 {
		return ""
	}
	switch vv := b.args[0].(type) {
	case string:
		return vv
	default:
		return fmt.Sprintf("%v", b.args[0])
	}
}

func (b *baseCmd) Args() []any {
	return b.args
}

func NewAnyCmd(args ...any) *AnyCmd {
	return &AnyCmd{
		baseCmd: baseCmd{args: args},
	}
}

var _ Cmder = (*AnyCmd)(nil)
var _ CmdResult = (*AnyCmd)(nil)

type AnyCmd struct {
	baseCmd
	Result
}

func convert[T any](result resp3.Element, err error) (v T, e error) {
	obj, err := resp3.ToAny(result, err)
	if err != nil {
		return v, err
	}
	mp, ok := obj.(T)
	if ok {
		return mp, nil
	}
	return v, fmt.Errorf("got type %T, expected %T", result, v)
}

func NewResult(reply resp3.Element, err error) *Result {
	ret := &Result{}
	ret.SetReply(reply, err)
	return ret
}

var _ CmdResult = (*Result)(nil)

// Result Cmd 执行后的结果
type Result struct {
	reply resp3.Element
	value any
	err   error
}

func (ac *Result) SetReply(result resp3.Element, err error) {
	ac.reply = result
	ac.err = err
	ac.value, ac.err = resp3.ToAny(result, err)
}

func (ac *Result) SetErr(err error) {
	ac.err = err
}

func (ac *Result) Err() error {
	return ac.err
}

func (ac *Result) ValueErr() (any, error) {
	return ac.value, ac.err
}

func (ac *Result) Value() any {
	return ac.value
}

func (ac *Result) Int() (int, error) {
	return resp3.ToInt(ac.reply, ac.err)
}

func (ac *Result) Int64() (int64, error) {
	return resp3.ToInt64(ac.reply, ac.err)
}

func (ac *Result) Float64() (float64, error) {
	return resp3.ToFloat64(ac.reply, ac.err)
}

func (ac *Result) Float64Slice() ([]float64, error) {
	return resp3.ToFloat64Slice(ac.reply, ac.err)
}

func (ac *Result) String() (string, error) {
	return resp3.ToString(ac.reply, ac.err)
}

func (ac *Result) StringSlice() ([]string, error) {
	return resp3.ToStringSlice(ac.reply, ac.err)
}

func (ac *Result) OKStatus() error {
	return resp3.ToOkStatus(ac.reply, ac.err)
}

func (ac *Result) StringMap() (map[string]string, error) {
	return resp3.ToStringMap(ac.reply, ac.err)
}

func (ac *Result) StringAnyMap() (map[string]any, error) {
	return convert[map[string]any](ac.reply, ac.err)
}

func (ac *Result) Map() (map[any]any, error) {
	return resp3.ToAnyMap(ac.reply, ac.err)
}

func (ac *Result) Slice() ([]any, error) {
	return resp3.ToAnySlice(ac.reply, ac.err)
}
