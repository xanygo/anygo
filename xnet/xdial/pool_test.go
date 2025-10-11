//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-11

package xdial_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/iotest"

	"github.com/fsgo/fst"

	"github.com/xanygo/anygo/ds/xpool"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xdial"
	"github.com/xanygo/anygo/xoption"
)

func Test_ConnPool1(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello:" + req.RemoteAddr))
	}))
	addr := ts.Listener.Addr()
	cc := xdial.ConnectorFunc(func(ctx context.Context, _ xnet.AddrNode, opt xoption.Reader) (*xnet.ConnNode, error) {
		a := xnet.AddrNode{
			Addr: addr,
		}
		return xdial.Connect(ctx, nil, a, opt)
	})
	pool, err1 := xdial.NewGroupPool("long", &xpool.Option{}, cc)
	fst.NoError(t, err1)
	fst.NotEmpty(t, pool)
	defer pool.Close()

	hc := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				ad := xnet.AddrNode{
					Addr: xnet.NewAddr(network, addr),
				}
				ent, err := pool.Get(ctx, ad)
				if err != nil {
					return nil, err
				}
				conn := ent.Object()
				conn.OnClose = func() error {
					conn.OnClose = nil
					ent.Release(conn.Err())
					return nil
				}
				return conn, nil
			},
		},
	}
	resp2, err2 := hc.Get(ts.URL)
	fst.NoError(t, err2)
	content3, err3 := io.ReadAll(resp2.Body)
	fst.NoError(t, resp2.Body.Close())
	fst.NoError(t, err3)

	for i := 0; i < 100; i++ {
		t.Run(fmt.Sprintf("loop_%d", i), func(t *testing.T) {
			resp4, err4 := hc.Get(ts.URL)
			fst.NoError(t, err4)
			// 验证client使用的是同一个连接
			fst.NoError(t, iotest.TestReader(resp4.Body, content3))
			fst.NoError(t, resp2.Body.Close())
		})
	}
}
