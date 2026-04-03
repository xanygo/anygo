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
	return context.Cause(ctx)
}

var ctxKeyClientConn = NewKey()

// WithClientConn 用于 server 的 handler 中存储用于读写的文件句柄
func WithClientConn[C any](ctx context.Context, conn C) context.Context {
	return context.WithValue(ctx, ctxKeyClientConn, conn)
}

func ClientConn[C any](ctx context.Context) C {
	val, _ := ctx.Value(ctxKeyClientConn).(C)
	return val
}
