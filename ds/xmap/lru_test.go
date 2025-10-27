//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-27

package xmap_test

import (
	"testing"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/xt"
)

func TestLRU(t *testing.T) {
	lru := xmap.NewLRU[string, string](128)
	lru.Set("k1", "v1")
	lru.Set("k2", "v2")
	lru.Set("k3", "v3")
	xt.Equal(t, []string{"k3", "k2", "k1"}, lru.Keys())

	got, ok := lru.Get("k2")
	xt.Equal(t, "v2", got)
	xt.Equal(t, true, ok)

	xt.Equal(t, []string{"k2", "k3", "k1"}, lru.Keys())
}
