//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-05-11

package xrpc_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/xanygo/anygo/xhttp/xhttpc"
	"github.com/xanygo/anygo/xnet/xrpc"
	"github.com/xanygo/anygo/xt"
)

func TestOptBlockPrivateIPs(t *testing.T) {
	body := []byte("Ok")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write(body)
	}))
	defer ts.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	t.Run("case 1 success", func(t *testing.T) {
		got, err := xhttpc.GetBody(ctx, "dummy", ts.URL)
		xt.NoError(t, err)
		xt.Equal(t, got, body)
	})
	t.Run("case 2 blocked", func(t *testing.T) {
		got, err := xhttpc.GetBody(ctx, "dummy", ts.URL, xrpc.OptBlockPrivateIPs())
		xt.Error(t, err)
		xt.Empty(t, got)
	})
}
