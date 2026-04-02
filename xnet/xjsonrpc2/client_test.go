//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-04-01

package xjsonrpc2_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/xanygo/anygo/ds/xoption"
	"github.com/xanygo/anygo/xnet/dsession"
	"github.com/xanygo/anygo/xnet/xjsonrpc2"
	"github.com/xanygo/anygo/xnet/xrpc"
	"github.com/xanygo/anygo/xnet/xservice"
	"github.com/xanygo/anygo/xt"
)

func pingHandler(ctx context.Context, req *xjsonrpc2.Request) (result any, err error) {
	var payload string
	err = req.DecodeParams(&payload)
	if err != nil {
		return nil, err
	}
	return "Ok: " + payload, nil
}

func TestClientRequest1(t *testing.T) {
	router := xjsonrpc2.NewRouter()
	router.RegisterFunc("ping", pingHandler)
	ser := httptest.NewServer(router)
	defer ser.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	t.Run("case 1", func(t *testing.T) {
		srv, err := xservice.NewServiceByURL("test", ser.URL)
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
				err = xrpc.Invoke(ctx, srv, req, resp, xrpc.OptDialHandshake(ss))
				xt.Equal(t, "Ok: hello", resp.Result)
			})
		}
	})

	t.Run("case 2", func(t *testing.T) {
		srv, err := xservice.NewServiceByURL("test", ser.URL)
		xt.NoError(t, err)
		xt.NotNil(t, srv)

		// 添加配置，使其自动的完成协议升级转换
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

		for i := 1; i < 3; i++ {
			t.Run(fmt.Sprintf("loop_%d", i), func(t *testing.T) {
				req := &xjsonrpc2.ClientRequest[string]{
					ID:     xjsonrpc2.Int64ID(i),
					Method: "ping",
					Params: "hello",
				}
				resp := &xjsonrpc2.ClientResponse[string]{}
				err = xrpc.Invoke(ctx, srv, req, resp)
				xt.Equal(t, "Ok: hello", resp.Result)
			})
		}
	})
}
