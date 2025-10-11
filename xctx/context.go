//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-01

package xctx

import (
	"context"
	"sync/atomic"
)

type Key struct {
	id int64
}

func (k *Key) ID() int64 {
	return k.id
}

var keyID atomic.Int64

// NewKey 一个新的 context Key，一般创建后赋值给一个全局变量，而不是每次调用都创建一个新的
func NewKey() *Key {
	return &Key{
		id: keyID.Add(1),
	}
}

// CheckError 若 err!=nil 则返回 error，否则返回 ctx.Err()
func CheckError(ctx context.Context, err error) error {
	if err != nil {
		return err
	}
	return ctx.Err()
}

var signKey = NewKey()

func WithSign(ctx context.Context, sign string) context.Context {
	return WithValues(ctx, signKey, sign)
}

func Signs(ctx context.Context) []string {
	return Values[*Key, string](ctx, signKey, true)
}
