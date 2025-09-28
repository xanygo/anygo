//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-27

package xhttpc_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/fsgo/fst"

	"github.com/xanygo/anygo/store/xcache"
	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xhttp/xhttpc"
	"github.com/xanygo/anygo/xnet/xservice"
)

func TestCacheClient(t *testing.T) {
	var id atomic.Int64
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Ok", "Ok")
		w.Write([]byte(strconv.FormatInt(id.Add(1), 10)))
	}))
	defer ts.Close()

	xservice.DefaultRegistry().Register(xservice.DummyService())

	rc := xcache.NewLRU[string, *xhttpc.StoredResponse](10)
	hc := xhttpc.CacheClient{
		Cache:   rc,
		Request: httptest.NewRequest(http.MethodGet, ts.URL, nil),
		Decoder: xcodec.Raw,
	}

	for i := 0; i < 100; i++ {
		t.Run(fmt.Sprintf("i_%d", i), func(t *testing.T) {
			resp := &xhttpc.StoredResponse{}
			err := hc.Invoke(context.Background(), resp)
			fst.NoError(t, err)
			fst.Equal(t, "1", string(resp.Body))
			fst.Equal(t, "Ok", resp.Header.Get("X-Ok"))
			fst.Greater(t, resp.CreateAt, 1)
			fst.Equal(t, 200, resp.StatusCode)
			if i == 0 {
				fst.False(t, resp.FromCache)
			} else {
				fst.True(t, resp.FromCache)
			}
			fst.Equal(t, 1, id.Load())
		})
	}

	synctest.Test(t, func(t *testing.T) {
		hc.TTL = time.Hour
		hc.PreFlush = 1000 * time.Second
		time.Sleep(1001 * time.Second)

		resp1 := &xhttpc.StoredResponse{}
		err1 := hc.Invoke(context.Background(), resp1)
		fst.NoError(t, err1)
		fst.Equal(t, "1", string(resp1.Body))
		fst.Equal(t, "Ok", resp1.Header.Get("X-Ok"))
		fst.Greater(t, resp1.CreateAt, 1)
		fst.Equal(t, 200, resp1.StatusCode)
		fst.True(t, resp1.FromCache)
	})
}
