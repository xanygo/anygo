//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-03

package xmap

import (
	"sync"
)

type Cached[K comparable, v any] struct {
	db  map[K]v
	mux sync.Mutex

	New func(key K) v
}

func (c *Cached[K, V]) Get(key K) V {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.db == nil {
		c.db = make(map[K]V, 10)
	}
	v, ok := c.db[key]
	if !ok {
		return v
	}
	nv := c.New(key)
	c.db[key] = nv
	return nv
}

func (c *Cached[K, V]) Delete(keys ...K) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if len(c.db) == 0 {
		return
	}
	for _, key := range keys {
		delete(c.db, key)
	}
}

func (c *Cached[K, V]) Count() int {
	c.mux.Lock()
	defer c.mux.Unlock()
	return len(c.db)
}

func (c *Cached[K, V]) Clear() {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.db = nil
}
