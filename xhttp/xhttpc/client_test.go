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

	"github.com/xanygo/anygo/ds/xctx"
	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/ds/xurl"
	"github.com/xanygo/anygo/store/xcache"
	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xhttp/xhttpc"
	"github.com/xanygo/anygo/xlog"
	"github.com/xanygo/anygo/xnet/xrpc"
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
	hc := xhttpc.CachedClient{
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
		n, err := io.Copy(w, r.Body)
		t.Logf("HandlerFunc io.Copy %d,%v", n, err)
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

func TestClientRetry(t *testing.T) {
	var requestID atomic.Int64
	var handlerStore xsync.Value[http.HandlerFunc]
	defaultHandler := func(w http.ResponseWriter, r *http.Request) {
		rid := requestID.Load()
		t.Logf("requestID=%d method=%s logID=%s retryCount=%s uri=%s",
			rid, r.Method, r.Header.Get("X-Log-ID"), r.Header.Get("X-Retry-Count"), r.RequestURI)
		query := r.URL.Query()
		if rid == 1 {
			if query.Has("sleep") {
				xctx.Sleep(r.Context(), time.Second)
				return
			}
			status := xurl.IntDef(query, "code", 200)
			w.WriteHeader(status)
		} else {
			w.WriteHeader(200)
		}
		body := xurl.StringDef(query, "body", "Ok")
		n, err := w.Write([]byte(body))
		t.Logf("ResponseWriter, n=%d, err=%v", n, err)
		xt.NoError(t, err)
	}
	handlerStore.Store(defaultHandler)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID.Add(1)
		handlerStore.Load().ServeHTTP(w, r)
	}))
	defer ts.Close()

	t.Run("get-ok", func(t *testing.T) {
		requestID.Store(0)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		ctx = xlog.NewContext(ctx)
		xlog.WithLogID(ctx, xlog.NewLogID())
		defer cancel()
		resp := &http.Response{}
		err := xhttpc.Get(ctx, "dummy", ts.URL, xhttpc.FetchResponse(resp), xrpc.OptRetry(2))
		xt.NoError(t, err)
		xt.Equal(t, 200, resp.StatusCode)
		xt.NoError(t, iotest.TestReader(resp.Body, []byte(`Ok`)))
		xt.Equal(t, 1, requestID.Load())
	})

	t.Run("get-timeout-retry", func(t *testing.T) {
		requestID.Store(0)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		ctx = xlog.NewContext(ctx)
		xlog.WithLogID(ctx, xlog.NewLogID())
		defer cancel()
		resp := &http.Response{}
		err := xhttpc.Get(ctx, "dummy", ts.URL+"?sleep=get-timeout-retry", xhttpc.FetchResponse(resp),
			xrpc.OptRetry(2), xrpc.OptReadTimeout(10*time.Millisecond))
		xt.NoError(t, err)
		xt.Equal(t, 200, resp.StatusCode)
		xt.NoError(t, iotest.TestReader(resp.Body, []byte(`Ok`)))
		xt.Equal(t, 2, requestID.Load()) // 值是 2，说明重试成功
	})

	t.Run("Post-timeout-always-retry", func(t *testing.T) {
		requestID.Store(0)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		ctx = xlog.NewContext(ctx)
		xlog.WithLogID(ctx, xlog.NewLogID())
		defer cancel()
		resp := &http.Response{}
		err := xhttpc.PostJSON(ctx, "dummy", ts.URL+"?sleep=Post-timeout-always-retry", map[string]any{},
			xhttpc.FetchResponse(resp), xrpc.OptRetry(2), xrpc.OptReadTimeout(20*time.Millisecond))
		xt.NoError(t, err)
		xt.Equal(t, 200, resp.StatusCode)
		xt.NoError(t, iotest.TestReader(resp.Body, []byte(`Ok`)))
		xt.Equal(t, 2, requestID.Load()) // 值是 2，说明重试成功
	})

	t.Run("Post-timeout-no-retry", func(t *testing.T) {
		// 默认重试策略：Post timeout 错误，不能重试

		requestID.Store(0)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		ctx = xlog.NewContext(ctx)
		xlog.WithLogID(ctx, xlog.NewLogID())
		defer cancel()
		resp := &http.Response{}
		start := time.Now()
		err := xhttpc.PostJSON(ctx, "dummy", ts.URL+"?sleep=Post-timeout-no-retry", map[string]any{},
			xhttpc.FetchResponse(resp), xrpc.OptRetryWithPolicy(2, nil), xrpc.OptReadTimeout(10*time.Millisecond))
		cost := time.Since(start)
		t.Logf("cost=%s", cost.String())
		xt.Equal(t, 1, requestID.Load()) // 值是 1，说明没有重试
		xt.Error(t, err)
		xt.Equal(t, 0, resp.StatusCode)
	})
}
