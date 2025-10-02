//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-30

package xredis

import (
	"bufio"
	"context"

	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/store/xredis/resp3"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xrpc"
	"github.com/xanygo/anygo/xnet/xservice"
	"github.com/xanygo/anygo/xoption"
)

type (
	Request = resp3.Request

	Result = resp3.Result
)

var ErrNil = resp3.ErrNil

func NewClient(service string) *Client {
	c := &Client{
		Service: service,
	}
	c.once = &xsync.OnceInit[[]xrpc.Option]{
		New: c.geRPCOptions,
	}
	return c
}

type Client struct {
	Service  any
	Registry xservice.Registry
	once     *xsync.OnceInit[[]xrpc.Option]
}

func (c *Client) geRPCOptions() []xrpc.Option {
	options := make([]xrpc.Option, 1, 2)
	options[0] = xrpc.OptHandshakeHandler(xrpc.HandshakeFunc(handshake))
	if c.Registry != nil {
		options = append(options, xrpc.OptServiceRegistry(c.Registry))
	}
	return options
}

func (c *Client) do(ctx context.Context, cmd Request) *rpcResponse {
	req := &rpcRequest{
		req: cmd,
	}
	resp := &rpcResponse{}
	_ = xrpc.Invoke(ctx, c.Service, req, resp, c.once.Load()...)
	return resp
}

func (c *Client) Do(ctx context.Context, cmd Request) Response {
	return Response{
		base: c.do(ctx, cmd),
	}
}

func handshake(ctx context.Context, conn *xnet.ConnNode, opt xoption.Reader) (any, error) {
	cmd := resp3.HelloRequest{}
	bf := bp.Get()
	_, err := conn.Conn.Write(cmd.Bytes(bf))
	bp.Put(bf)
	if err != nil {
		return nil, err
	}
	br := bufio.NewReader(conn.Conn)
	return resp3.ReadByType(br, cmd.ResponseType())
}
