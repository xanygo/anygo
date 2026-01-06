//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-27

package xhttpc_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"testing/iotest"
	"testing/synctest"
	"time"

	"github.com/xanygo/anygo/store/xcache"
	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xhttp/xhttpc"
	"github.com/xanygo/anygo/xnet/xservice"
	"github.com/xanygo/anygo/xt"
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
			xt.NoError(t, err)
			xt.Equal(t, "1", string(resp.Body))
			xt.Equal(t, "Ok", resp.Header.Get("X-Ok"))
			xt.Greater(t, resp.CreateAt, 1)
			xt.Equal(t, 200, resp.StatusCode)
			if i == 0 {
				xt.False(t, resp.FromCache)
			} else {
				xt.True(t, resp.FromCache)
			}
			xt.Equal(t, 1, id.Load())
		})
	}

	synctest.Test(t, func(t *testing.T) {
		hc.TTL = time.Hour
		hc.PreFlush = 1000 * time.Second
		time.Sleep(1001 * time.Second)

		resp1 := &xhttpc.StoredResponse{}
		err1 := hc.Invoke(context.Background(), resp1)
		xt.NoError(t, err1)
		xt.Equal(t, "1", string(resp1.Body))
		xt.Equal(t, "Ok", resp1.Header.Get("X-Ok"))
		xt.Greater(t, resp1.CreateAt, 1)
		xt.Equal(t, 200, resp1.StatusCode)
		xt.True(t, resp1.FromCache)
	})
}

func TestClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Ok", "Ok")
		defer r.Body.Close()
		io.Copy(w, r.Body)
	}))
	defer ts.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	t.Run("Get", func(t *testing.T) {
		resp := &xhttpc.StoredResponse{}
		err := xhttpc.Get(ctx, xservice.Dummy, ts.URL, xhttpc.TeeReader(resp))
		xt.NoError(t, err)
		xt.Equal(t, "Ok", resp.Header.Get("X-Ok"))
	})

	t.Run("PostJSON", func(t *testing.T) {
		resp := &xhttpc.StoredResponse{}
		data := map[string]any{
			"k1": "v1",
		}
		err := xhttpc.PostJSON(ctx, xservice.Dummy, ts.URL, data, xhttpc.TeeReader(resp))
		xt.NoError(t, err)
		xt.Equal(t, "Ok", resp.Header.Get("X-Ok"))
		xt.Equal(t, `{"k1":"v1"}`, string(resp.Body))
	})

	t.Run("Client.PostJSON", func(t *testing.T) {
		data := map[string]any{
			"k1": "v1",
		}
		c1 := &xhttpc.Client{}
		resp, err := c1.PostJSON(ctx, ts.URL, data)
		xt.NoError(t, err)
		xt.Equal(t, "Ok", resp.Header.Get("X-Ok"))
		defer resp.Body.Close()
		xt.NoError(t, iotest.TestReader(resp.Body, []byte(`{"k1":"v1"}`)))
	})
}
