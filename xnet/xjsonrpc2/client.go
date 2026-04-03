//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-03-30

package xjsonrpc2

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/xanygo/anygo/ds/xoption"
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xio"
	"github.com/xanygo/anygo/xnet/xrpc"
)

// 使用 xrpc.Client 发送请求和信息的实现

var _ xrpc.Request = (*ClientRequest[any])(nil)
var _ noReply = (*ClientRequest[any])(nil)

type ClientRequest[P any] struct {
	// ID 客户端的唯一标识id，值必须包含一个字符串、数值或 NULL 空值
	ID ID

	// Method 方法名
	Method string

	// Params 请求参数
	Params P
}

// NoReply 是否是通知/不需要回复
func (c *ClientRequest[P]) NoReply() bool {
	return c.ID == nil
}

func (c *ClientRequest[P]) String() string {
	return fmt.Sprintf("%s ClientRequest method=%q id=%q", Protocol, c.Method, idBytes(c.ID))
}

func (c *ClientRequest[P]) Protocol() string {
	return Protocol
}

func (c *ClientRequest[P]) APIName() string {
	return c.Method
}

func (c *ClientRequest[P]) WriteTo(ctx context.Context, w io.Writer, opt xoption.Reader) error {
	rr := &Request{
		ID:     c.ID,
		Method: c.Method,
	}
	err := rr.WithParams(c.Params)
	if err != nil {
		return err
	}

	if ds, ok := w.(xio.WriteDeadlineSetter); ok {
		timeout := xoption.WriteTimeout(opt)
		if err = ds.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
			return err
		}
		defer ds.SetWriteDeadline(time.Time{})
	}

	_, err = rr.WriteTo(w)
	return err
}

var _ xrpc.Response = (*ClientResponse[any])(nil)

type ClientResponse[P any] struct {
	ID     ID
	Error  *Error
	Result P
	raw    *Response
}

func (c *ClientResponse[P]) String() string {
	return fmt.Sprintf("%s ClientResponse ID=%s, Error=%v", Protocol, idBytes(c.ID), c.Error)
}

func (c *ClientResponse[P]) LoadFrom(_ context.Context, req xrpc.Request, r io.Reader, opt xoption.Reader) error {
	c.raw = nil
	if nr, ok := req.(noReply); ok && nr.NoReply() {
		return nil
	}
	if ds, ok := r.(xio.ReadDeadlineSetter); ok {
		timeout := xoption.ReadTimeout(opt)
		if err := ds.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			return err
		}
		defer ds.SetReadDeadline(time.Time{})
	}

	maxSize := xoption.MaxResponseSize(opt)
	bio := bufio.NewReader(io.LimitReader(r, maxSize))

	resp, err := ReadResponse(bio)
	if err != nil {
		return err
	}
	c.raw = resp
	return resp.DecodeResult(&c.Result)
}

func (c *ClientResponse[P]) ErrCode() int64 {
	if c.Error == nil {
		return 0
	}
	return c.Error.ErrCode()
}

func (c *ClientResponse[P]) ErrMsg() string {
	if c.Error == nil {
		return ""
	}
	return c.Error.Message
}

func (c *ClientResponse[P]) Unwrap() any {
	return c.raw
}

func (c *ClientResponse[P]) RawResult() json.RawMessage {
	if c.raw == nil {
		return nil
	}
	return c.raw.Result
}

// 下面是批量的实现
var _ xrpc.Request = ClientRequests[any]{}

// ClientRequests 用于发送批量请求
type ClientRequests[P any] []ClientRequest[P]

func (crs ClientRequests[P]) String() string {
	return fmt.Sprintf("ClientRequests, len=%d", len(crs))
}

func (crs ClientRequests[P]) Protocol() string {
	return Protocol
}

func (crs ClientRequests[P]) APIName() string {
	return "batch"
}

func (crs ClientRequests[P]) WriteTo(_ context.Context, w io.Writer, opt xoption.Reader) error {
	if len(crs) == 0 {
		return errors.New("empty request")
	}
	bf := bytes.NewBuffer(nil)
	bf.WriteString("[")
	for index, cr := range crs {
		rr := &Request{
			ID:     cr.ID,
			Method: cr.Method,
		}
		err := rr.WithParams(cr.Params)
		if err == nil {
			_, err = rr.WriteTo(bf)
		}
		if err != nil {
			return err
		}
		if index < len(crs)-1 {
			bf.WriteString(",")
		}
	}
	bf.WriteString("]\n")

	if ds, ok := w.(xio.WriteDeadlineSetter); ok {
		timeout := xoption.WriteTimeout(opt)
		if err := ds.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
			return err
		}
		defer ds.SetWriteDeadline(time.Time{})
	}

	_, err := w.Write(bf.Bytes())
	return err
}

func (crs ClientRequests[P]) NoReply() bool {
	for _, cr := range crs {
		if !cr.NoReply() {
			return false
		}
	}
	return true
}

var _ xrpc.Response = (*ClientResponses)(nil)

// ClientResponses  和 ClientRequests 匹配对应的响应解析逻辑
type ClientResponses struct {
	values []*Response
	err    error
}

func (c *ClientResponses) String() string {
	return "ClientResponses"
}

func (c *ClientResponses) LoadFrom(_ context.Context, req xrpc.Request, r io.Reader, opt xoption.Reader) error {
	c.values = nil
	c.err = nil
	if nr, ok := req.(noReply); ok && nr.NoReply() {
		return nil
	}
	if ds, ok := r.(xio.ReadDeadlineSetter); ok {
		timeout := xoption.ReadTimeout(opt)
		if err := ds.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			return err
		}
		defer ds.SetReadDeadline(time.Time{})
	}
	bio := bufio.NewReader(io.LimitReader(r, xoption.MaxResponseSize(opt)))
	c.values, _, c.err = ReadResponses(bio)
	return c.err
}

func (c *ClientResponses) ErrCode() int64 {
	if c.err == nil {
		return 0
	}
	return xerror.ErrCode(c.err, 1)
}

func (c *ClientResponses) ErrMsg() string {
	if c.err == nil {
		return ""
	}
	return c.err.Error()
}

func (c *ClientResponses) Result() []*Response {
	return c.values
}

func (c *ClientResponses) Unwrap() any {
	return c.values
}
