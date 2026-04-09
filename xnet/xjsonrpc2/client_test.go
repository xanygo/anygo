//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-04-01

package xjsonrpc2_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/xanygo/anygo/ds/xctx"
	"github.com/xanygo/anygo/ds/xoption"
	"github.com/xanygo/anygo/xnet/dsession"
	"github.com/xanygo/anygo/xnet/internal"
	"github.com/xanygo/anygo/xnet/xdial"
	"github.com/xanygo/anygo/xnet/xjsonrpc2"
	"github.com/xanygo/anygo/xnet/xrpc"
	"github.com/xanygo/anygo/xnet/xservice"
	"github.com/xanygo/anygo/xt"
)

func pingHandler(ctx context.Context, req *xjsonrpc2.Request) (result any, err error) {
	var payload string
	err = req.DecodeParams(&payload)
	log.Println("call pingHandler, id=", req.ID, "payload=", payload)
	if err != nil {
		return nil, err
	}
	return "Ok: " + payload, nil
}

func infoHandler(ctx context.Context, req *xjsonrpc2.Request) (result any, err error) {
	log.Println("call infoHandler")
	cc := xctx.ClientConn[net.Conn](ctx)
	if cc == nil {
		return nil, errors.New("invalid client conn")
	}
	result = fmt.Sprintf("id=%v, client=%s", req.ID, cc.LocalAddr().String())
	return result, nil
}

func TestClientRequest1(t *testing.T) {
	router := xjsonrpc2.NewRouter()
	router.RegisterUnary("ping", pingHandler)
	router.RegisterUnary("info", infoHandler)
	ts := httptest.NewServer(router)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	t.Run("case 1 over http", func(t *testing.T) {
		srv, err := xservice.NewServiceByURL("test", ts.URL)
		xt.NoError(t, err)
		xt.NotNil(t, srv)

		xt.NoError(t, srv.Start(ctx))
		defer srv.Stop(ctx)

		// 手工协议升级
		ss := dsession.HTTPUpgrade(http.MethodPost, "/", "json-rpc2")
		for i := 1; i < 10; i++ {
			t.Run(fmt.Sprintf("loop_%d", i), func(t *testing.T) {
				req := &xjsonrpc2.ClientRequest[string]{
					ID:     xjsonrpc2.Int64ID(i),
					Method: "ping",
					Params: "hello",
				}
				resp := &xjsonrpc2.ClientResponse[string]{}
				err = xrpc.Invoke(ctx, srv, req, resp, xrpc.OptSessionInit(ss))
				xt.Equal(t, resp.Result, "Ok: hello")
			})
		}
	})

	t.Run("case 2 auto http upgrade with short conn", func(t *testing.T) {
		srv, err := xservice.NewServiceByURL("test", ts.URL)
		xt.NoError(t, err)
		xt.NotNil(t, srv)

		// 添加配置，使其自动完成协议升级转换
		optw, ok := srv.Option().(xoption.Writer)
		xt.True(t, ok)
		xoption.SetSessionStarter(optw, &xoption.SessionStarterConfig{
			Name: "HTTP-Upgrade",
			Params: map[string]any{
				"Method":   http.MethodPost,
				"URI":      "/api",
				"Protocol": xjsonrpc2.Protocol,
			},
		})

		xt.NoError(t, srv.Start(ctx))
		defer srv.Stop(ctx)

		t.Run("ping-pong", func(t *testing.T) {
			for i := 1; i < 3; i++ {
				t.Run(fmt.Sprintf("loop_%d", i), func(t *testing.T) {
					req := &xjsonrpc2.ClientRequest[string]{
						ID:     xjsonrpc2.Int64ID(i),
						Method: "ping",
						Params: "hello",
					}
					resp := &xjsonrpc2.ClientResponse[string]{}
					err = xrpc.Invoke(ctx, srv, req, resp)
					xt.Equal(t, resp.Result, "Ok: hello")
				})
			}
		})
	})

	t.Run("case 3 auto http upgrade with long conn", func(t *testing.T) {
		address, err := internal.HostPortFromURL(ts.URL)
		xt.NoError(t, err)
		cfg := &xservice.Config{
			Name: "test",
			ConnPool: &xservice.ConnPoolPart{
				Name: xdial.Long,
			},
			SessionInit: &xoption.SessionStarterConfig{
				Name: "HTTP-Upgrade",
				Params: map[string]any{
					"Method":   http.MethodPost,
					"URI":      "/api",
					"Protocol": xjsonrpc2.Protocol,
				},
			},
			DownStream: xservice.DownStreamPart{
				Address: []string{address},
			},
		}
		srv, err := cfg.Parser("test")
		xt.NoError(t, err)
		xt.NotNil(t, srv)

		xt.NoError(t, srv.Start(ctx))
		defer srv.Stop(ctx)

		t.Run("ping-pong", func(t *testing.T) {
			for i := 1; i < 3; i++ {
				t.Run(fmt.Sprintf("loop_%d", i), func(t *testing.T) {
					req := &xjsonrpc2.ClientRequest[string]{
						ID:     xjsonrpc2.Int64ID(i),
						Method: "ping",
						Params: "hello",
					}
					resp := &xjsonrpc2.ClientResponse[string]{}
					err = xrpc.Invoke(ctx, srv, req, resp)
					xt.Equal(t, resp.Result, "Ok: hello")
				})
			}
		})

		t.Run("info", func(t *testing.T) {
			req := &xjsonrpc2.ClientRequest[string]{
				ID:     xjsonrpc2.Int64ID(1),
				Method: "info",
				Params: "hello",
			}
			resp := &xjsonrpc2.ClientResponse[string]{}
			err = xrpc.Invoke(ctx, srv, req, resp)
			xt.NoError(t, err)
			ret1 := resp.Result
			xt.HasPrefix(t, ret1, "id=1, client=")

			err = xrpc.Invoke(ctx, srv, req, resp)
			xt.NoError(t, err)
			xt.Equal(t, resp.Result, ret1)
		})
	})
}

func TestClient(t *testing.T) {
	router := xjsonrpc2.NewRouter()
	router.RegisterUnary("ping", pingHandler)
	router.RegisterUnary("info", infoHandler)
	ts := httptest.NewServer(router)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	srv, err := xservice.NewServiceByURL("test", ts.URL)
	xt.NoError(t, err)
	xt.NotNil(t, srv)
	xt.NoError(t, srv.Start(ctx))
	defer srv.Stop(ctx)

	ss := dsession.HTTPUpgrade(http.MethodPost, "/", "json-rpc2")
	client := &xjsonrpc2.Client{
		Service: srv,
		RPCOptions: []xrpc.Option{
			xrpc.OptSessionInit(ss),
		},
	}
	t.Run("InvokeUnary", func(t *testing.T) {
		req, _ := xjsonrpc2.NewRequest(xjsonrpc2.Int64ID(1), "ping", "hello")
		resp, err := client.InvokeUnary(ctx, req)
		xt.NoError(t, err)
		var str string
		xt.NoError(t, resp.DecodeResult(&str))
		xt.Equal(t, str, "Ok: hello")
	})

	t.Run("Stream", func(t *testing.T) {
		ctx, cancel = context.WithCancel(ctx)
		defer cancel()

		sender := make(chan *xjsonrpc2.Request)
		rec, err := client.Stream(ctx, sender)
		xt.NoError(t, err)
		xt.NotNil(t, rec)
		go func() {
			for i := 1; i <= 10; i++ {
				req, _ := xjsonrpc2.NewRequest(xjsonrpc2.Int64ID(i), "ping", "hello")
				sender <- req
			}
			close(sender)
		}()
		resps := make([]*xjsonrpc2.Response, 0)

	Loop:
		for i := 1; i <= 10; i++ {
			select {
			case <-ctx.Done():
				break Loop
			case resp, ok := <-rec:
				if !ok {
					break Loop
				}
				var str string
				xt.NoError(t, resp.DecodeResult(&str))
				xt.Equal(t, str, "Ok: hello")
				resps = append(resps, resp)
			}
		}
		xt.Len(t, resps, 10)
	})
}
