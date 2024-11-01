//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-01

package xctx

import "sync/atomic"

type Key struct {
	id int32
}

var keyID atomic.Int32

func NewKey() *Key {
	return &Key{
		id: keyID.Add(1),
	}
}
