//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-01

package xredis

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/xanygo/anygo/ds/xoption"
	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/store/xredis/resp3"
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xrpc"
)

var _ xrpc.Request = (*rpcRequest)(nil)

type rpcRequest struct {
	req resp3.Request
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

	timeout := xoption.WriteTimeout(opt)
	if err := w.SetDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	defer w.SetDeadline(time.Time{})

	_, err := w.Write(content)
	bp.Put(bf)
	return err
}

var _ xrpc.Response = (*rpcResponse)(nil)

type rpcResponse struct {
	result resp3.Element
	err    error
}

func (resp *rpcResponse) String() string {
	if resp.err == nil {
		return fmt.Sprintf("redis resullt: %T", resp.result)
	}
	return resp.err.Error()
}

func (resp *rpcResponse) LoadFrom(ctx context.Context, req xrpc.Request, rd *xnet.ConnNode, opt xoption.Reader) error {
	xrr, ok := req.(*rpcRequest)
	if !ok {
		return errors.New("not a redis rpcRequest")
	}

	timeout := xoption.ReadTimeout(opt)
	if err := rd.SetDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	defer rd.SetDeadline(time.Time{})

	br := bufio.NewReader(io.LimitReader(rd, xoption.MaxResponseSize(opt)))
	resp.result, resp.err = resp3.ReadByType(br, xrr.req.ResponseType())

	// 不需要将此错误返回给 xrpc.Client
	if errors.Is(resp.err, ErrNil) {
		return nil
	}
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

func (resp *rpcResponse) asResp3Array(count int) (resp3.Array, error) {
	if resp.err != nil {
		return nil, resp.err
	}
	if resp.result.DataType() == resp3.DataTypeNull {
		return nil, ErrNil
	}
	arr, ok := resp.result.(resp3.Array)
	if !ok {
		return nil, fmt.Errorf("unexpected type %T", resp.result)
	}
	if count > 0 && len(arr) != count {
		return nil, fmt.Errorf("expected %d arrays, got %d", count, len(arr))
	}
	return arr, nil
}
