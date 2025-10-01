//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-30

package xredis

import (
	"context"

	"github.com/xanygo/anygo/store/xredis/resp3"
	"github.com/xanygo/anygo/xnet/xrpc"
	"github.com/xanygo/anygo/xnet/xservice"
)

type (
	Request = resp3.Request

	Result = resp3.Result
)

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
	_ = xrpc.Invoke(ctx, c.Service, req, resp)
	return resp
}

func (c *Client) Do(ctx context.Context, cmd Request) Response {
	return Response{
		base: c.do(ctx, cmd),
	}
}
