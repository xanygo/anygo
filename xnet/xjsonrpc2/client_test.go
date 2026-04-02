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

	"github.com/xanygo/anygo/xnet/xdial"
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

	srv, err := xservice.NewServiceByURL("test", ser.URL)
	xt.NoError(t, err)
	xt.NotNil(t, srv)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	xt.NoError(t, srv.Start(ctx))
	defer srv.Stop(ctx)

	t.Run("case 1", func(t *testing.T) {
		ss := xdial.HTTPUpgrade(http.MethodPost, "/", "json-rpc2")
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
}
