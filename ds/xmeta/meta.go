//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-03-31

package xmeta

import (
	"fmt"
	"sync/atomic"
)

type Setter interface {
	SetMeta(key, value any)
}

type Getter interface {
	GetMeta(key any) (value any, ok bool)
}

var (
	// KeySessionReply RPC Client 握手信息存储时用
	KeySessionReply = NewKey("SessionReply")

	// KeyLongPool 标记当前连接是否是长连接
	KeyLongPool = NewKey("LongPool")
)

type Key struct {
	txt string
}

func (k *Key) String() string {
	return k.txt
}

var keyID atomic.Int64

func NewKey(name string) *Key {
	id := keyID.Add(1)
	return &Key{
		txt: fmt.Sprintf("%d%s", id, name),
	}
}

func TrySet(obj any, key any, meta any) {
	if mc, ok := obj.(Setter); ok {
		mc.SetMeta(key, meta)
	}
}

func TryGet(obj any, key any) any {
	if mc, ok := obj.(Getter); ok {
		val, _ := mc.GetMeta(key)
		return val
	}
	return nil
}

func HasKey(obj any, key any) bool {
	if mc, ok := obj.(Getter); ok {
		_, ok1 := mc.GetMeta(key)
		return ok1
	}
	return false
}
