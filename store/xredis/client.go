//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-30

package xredis

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/store/xredis/resp3"
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xrpc"
	"github.com/xanygo/anygo/xnet/xservice"
	"github.com/xanygo/anygo/xoption"
)

type (
	Request = resp3.Request

	Result = resp3.Result
)

type Client struct {
	Service  string
	Registry xservice.Registry
}

func (c *Client) Do(ctx context.Context, cmd Request) Response {
	req := &rpcRequest{
		req: cmd,
	}
	resp := &rpcResponse{}
	err := xrpc.Invoke(ctx, c.Service, req, resp)
	rr := Response{
		err: err,
	}
	if err == nil {
		rr.result = resp.result
		rr.err = resp.err
	}
	return rr
}

type Response struct {
	result Result
	err    error
}

func (rs Response) Result() (Result, error) {
	return rs.result, rs.err
}

func (rs Response) Err() error {
	return rs.err
}

func (rs Response) Int() (int, error) {
	if rs.err != nil {
		return 0, rs.err
	}
	switch dv := rs.result.(type) {
	case resp3.Integer:
		return dv.Int(), nil
	case resp3.BigNumber:
		return int(dv.BigInt().Int64()), nil
	case resp3.Null:
		return 0, nil
	case resp3.SimpleString:
		return strconv.Atoi(dv.String())
	case resp3.BulkString:
		return strconv.Atoi(dv.String())
	default:
		return 0, fmt.Errorf("unexpected response type: %T", dv)
	}
}

func NewClient(service string) *Client {
	return &Client{
		Service: service,
	}
}

var _ xrpc.Request = (*rpcRequest)(nil)

type rpcRequest struct {
	req Request
}

func (r *rpcRequest) String() string {
	return r.req.Name()
}

func (r *rpcRequest) Protocol() string {
	return "redis"
}

func (r *rpcRequest) APIName() string {
	return r.req.Name()
}

var bp = xsync.NewBytesBufferPool(1024)

func (r *rpcRequest) WriteTo(ctx context.Context, w *xnet.ConnNode, opt xoption.Reader) error {
	bf := bp.Get()
	content := r.req.Bytes(bf)
	_, err := w.Conn.Write(content)
	bp.Put(bf)
	return err
}

var _ xrpc.Response = (*rpcResponse)(nil)

type rpcResponse struct {
	result resp3.Result
	err    error
}

func (resp *rpcResponse) String() string {
	if resp.err == nil {
		return fmt.Sprintf("redis resullt: %T", resp.result)
	}
	return resp.err.Error()
}

func (resp *rpcResponse) LoadFrom(ctx context.Context, req xrpc.Request, rd io.Reader, opt xoption.Reader) error {
	xrr, ok := req.(*rpcRequest)
	if !ok {
		return errors.New("not a redis rpcRequest")
	}
	br := bufio.NewReader(rd)
	resp.result, resp.err = resp3.ReadByType(br, xrr.req.ResponseType())
	return resp.err
}

func (resp *rpcResponse) ErrCode() int64 {
	if resp.err == nil {
		return 0
	}
	return xerror.ErrCode(resp.err, 1)
}

func (resp *rpcResponse) ErrMsg() string {
	if resp.err == nil {
		return ""
	}
	return resp.err.Error()
}

func (resp *rpcResponse) Unwrap() any {
	return resp.result
}
