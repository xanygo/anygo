//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-30

package xredis

import (
	"bufio"
	"context"

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
	return &Client{
		Service: service,
	}
}

type Client struct {
	Service  string
	Registry xservice.Registry
}

func (c *Client) do(ctx context.Context, cmd Request) *rpcResponse {
	req := &rpcRequest{
		req: cmd,
	}
	resp := &rpcResponse{}
	_ = xrpc.Invoke(ctx, c.Service, req, resp, xrpc.OptHandshakeHandler(xrpc.HandshakeFunc(handshake)))
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
