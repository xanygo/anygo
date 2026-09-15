// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xcookiejar_test

import (
	"context"
	"net/http"

	"github.com/xanygo/anygo/xenc/xcodec"
	"github.com/xanygo/anygo/xhttp/xcookiejar"
	"github.com/xanygo/anygo/xkv"
	"github.com/xanygo/anygo/xkv/xkvx"
)

func ExampleJar() {
	db1 := xkvx.NewMemoryAny[xcookiejar.Entry](xcodec.JSON)
	db2 := xkvx.NewMemory()
	store := &xcookiejar.KV{
		EntryStore: func(key string) xkv.Hash[xcookiejar.Entry] {
			return db1.Hash(key)
		},
		MetaStore: func() xkv.ZSet[string] {
			return db2.ZSet("cookiejar-meta")
		},
	}

	jar := &xcookiejar.Jar{
		Storage: store,
	}

	httpCall := func(ctx context.Context, url string) error {
		client := &http.Client{
			// 使用 CookieJar。由于可能涉及 RPC 调用，所以需要将 context 传入
			Jar: jar.WithContext(ctx),
		}
		resp, err := client.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		// do something ...
		return nil
	}
	_ = httpCall
}
