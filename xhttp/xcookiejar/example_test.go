// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xcookiejar_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/xanygo/anygo/xenc/xcodec"
	"github.com/xanygo/anygo/xhttp/xcookiejar"
	"github.com/xanygo/anygo/xkv"
	"github.com/xanygo/anygo/xkv/xkvx"
)

func ExampleJar() {
	// Start a server to give us cookies.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie("Flavor"); err != nil {
			http.SetCookie(w, &http.Cookie{Name: "Flavor", Value: "Chocolate Chip"})
		} else {
			cookie.Value = "Oatmeal Raisin"
			http.SetCookie(w, cookie)
		}
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		log.Fatal(err)
	}

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
		PSList:  publicsuffix.List,
	}

	client := &http.Client{
		Jar: jar,
	}

	if _, err = client.Get(u.String()); err != nil {
		log.Fatal(err)
	}

	fmt.Println("After 1st request:")
	for _, cookie := range jar.Cookies(u) {
		fmt.Printf("  %s: %s\n", cookie.Name, cookie.Value)
	}

	if _, err = client.Get(u.String()); err != nil {
		log.Fatal(err)
	}

	fmt.Println("After 2nd request:")
	for _, cookie := range jar.Cookies(u) {
		fmt.Printf("  %s: %s\n", cookie.Name, cookie.Value)
	}
	// Output:
	// After 1st request:
	//   Flavor: Chocolate Chip
	// After 2nd request:
	//   Flavor: Oatmeal Raisin
}
