//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-30

package xsession

import (
	"time"

	"github.com/xanygo/anygo/store/xcache"
)

func NewMemoryStore(caption int, ttl time.Duration) *CacheStore {
	return &CacheStore{
		TTL:   ttl,
		Cache: xcache.NewLRU[string, string](caption),
	}
}
