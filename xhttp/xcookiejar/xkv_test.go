// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xcookiejar_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/xanygo/anygo/xenc/xcodec"
	"github.com/xanygo/anygo/xhttp/xcookiejar"
	"github.com/xanygo/anygo/xkv"
	"github.com/xanygo/anygo/xkv/xkvx"
	"github.com/xanygo/anygo/xt"
)

func TestKV(t *testing.T) {
	db1 := xkvx.NewMemoryAny[xcookiejar.Entry](xcodec.JSON)
	t.Run("case 1", func(t *testing.T) {
		kv := &xcookiejar.KV{
			EntryStore: func(key string) xkv.Hash[xcookiejar.Entry] {
				return db1.Hash(key)
			},
		}
		testStore(t, kv)
	})

	t.Run("case 2", func(t *testing.T) {
		db2 := xkvx.NewMemory()
		kv := &xcookiejar.KV{
			EntryStore: func(key string) xkv.Hash[xcookiejar.Entry] {
				return db1.Hash(key)
			},
			MetaStore: func() xkv.ZSet[string] {
				return db2.ZSet("meta")
			},
		}
		testStore(t, kv)
	})
}

func testStore(t *testing.T, store xcookiejar.Storage) {
	jar := &xcookiejar.Jar{
		Storage: store,
	}
	jar = jar.WithContext(t.Context())

	cookies := []*http.Cookie{
		{Name: "name1", Value: "value", Path: "/admin/"},
		{Name: "name2", Value: "value", Path: "/"},
		{Name: "name3", Value: "value", Path: "/"},
	}
	u := &url.URL{Scheme: "http", Host: "example.com", Path: "/"}
	err := jar.SetCookiesContext(t.Context(), u, cookies)
	xt.NoError(t, err)

	items, err := jar.CookiesContext(t.Context(), u)
	xt.NoError(t, err)
	xt.NotEmpty(t, items)
}
