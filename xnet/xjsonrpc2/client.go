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
	"slices"
	"time"

	"github.com/xanygo/anygo/ds/xoption"
	"github.com/xanygo/anygo/safely"
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
	ID       ID
	Error    *Error
	Result   P
	OnNotify func(resp *Response)
	raw      *Response
}

func (c *ClientResponse[P]) String() string {
	return fmt.Sprintf("%s ClientResponse ID=%s, Error=%v", Protocol, idBytes(c.ID), c.Error)
}

func (c *ClientResponse[P]) LoadFrom(_ context.Context, req xrpc.Request, r io.Reader, opt xoption.Reader) error {
	c.raw = nil
	if nr, ok := req.(noReply); ok && nr.NoReply() {
		return nil
	}

	var resp *Response
	var err error
	for {
		resp, err = c.readOne(r, opt)
		if err != nil {
			return err
		}
		if !resp.IsNotify() {
			break
		}
		if c.OnNotify != nil {
			c.OnNotify(resp)
		}
	}
	c.raw = resp
	return resp.DecodeResult(&c.Result)
}

func (c *ClientResponse[P]) readOne(r io.Reader, opt xoption.Reader) (*Response, error) {
	if ds, ok := r.(xio.ReadDeadlineSetter); ok {
		timeout := xoption.ReadTimeout(opt)
		if err := ds.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			return nil, err
		}
		defer ds.SetReadDeadline(time.Time{})
	}
	maxSize := xoption.MaxResponseSize(opt)
	bio := bufio.NewReader(io.LimitReader(r, maxSize))
	return ReadResponse(bio)
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

func (c *ClientResponse[P]) RawResponse() *Response {
	return c.raw
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

// Client 可以接收 server 推送消息的 Client
type Client struct {
	Service    any
	RPCOptions []xrpc.Option
}

func (c *Client) Clone() *Client {
	return &Client{
		Service:    c.Service,
		RPCOptions: slices.Clone(c.RPCOptions),
	}
}

// InvokeUnary 发送同步的一元请求
//
// 该方法不可以和 Stream 同时使用，否则收到的消息可能会出现混乱
func (c *Client) InvokeUnary(ctx context.Context, req *Request) (*Response, error) {
	creq := &ClientRequest[json.RawMessage]{
		ID:     req.ID,
		Method: req.Method,
		Params: req.Params,
	}
	cres := &ClientResponse[json.RawMessage]{}
	err := xrpc.Invoke(ctx, c.Service, creq, cres, c.RPCOptions...)
	if err != nil {
		return nil, err
	}
	return cres.RawResponse(), nil
}

// SendNotify 发送通知消息
func (c *Client) SendNotify(ctx context.Context, req *Request) error {
	if !req.NoReply() {
		return fmt.Errorf("has id=%v, not notify request", req.ID)
	}
	creq := &ClientRequest[json.RawMessage]{
		Method: req.Method,
		Params: req.Params,
	}
	return xrpc.Invoke(ctx, c.Service, creq, xrpc.NoResponse(), c.RPCOptions...)
}

// Stream 采用 stream 模式交互，异步的发送请求，异步读取响应
func (c *Client) Stream(ctx context.Context, sender <-chan *Request) (receiver <-chan *Response, err error) {
	req := &chatClientRequest{
		ch: sender,
	}
	rec := make(chan *Response, 32)
	resp := &chatClientResponse{
		ch: rec,
	}
	opts := slices.Clone(c.RPCOptions)
	opts = append(opts, xrpc.OptFullDuplex(true))
	go safely.Run(func() {
		defer close(rec)
		err = xrpc.Invoke(ctx, c.Service, req, resp, opts...)
	})
	return rec, err
}

var _ xrpc.Request = (*chatClientRequest)(nil)

type chatClientRequest struct {
	ch <-chan *Request
}

func (c *chatClientRequest) String() string {
	return "chatClientRequest"
}

func (c *chatClientRequest) Protocol() string {
	return Protocol
}

func (c *chatClientRequest) APIName() string {
	return ":stream"
}

func (c *chatClientRequest) WriteTo(ctx context.Context, w io.Writer, opt xoption.Reader) error {
	for {
		select {
		case req, ok := <-c.ch:
			if !ok {
				return nil
			}
			if err := c.sendOne(req, w, opt); err != nil {
				return err
			}
		case <-ctx.Done():
			return context.Cause(ctx)
		}
	}
}

func (c *chatClientRequest) sendOne(req *Request, w io.Writer, opt xoption.Reader) error {
	if ds, ok := w.(xio.WriteDeadlineSetter); ok {
		timeout := xoption.WriteTimeout(opt)
		if err := ds.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
			return err
		}
		defer ds.SetWriteDeadline(time.Time{})
	}
	return req.Write(w)
}

var _ xrpc.Response = (*chatClientResponse)(nil)

type chatClientResponse struct {
	ch  chan *Response
	err error
}

func (c *chatClientResponse) String() string {
	return "chatClientResponse"
}

func (c *chatClientResponse) LoadFrom(ctx context.Context, req xrpc.Request, r io.Reader, opt xoption.Reader) error {
	go safely.Run(func() {
		<-ctx.Done()
		if rc, ok := r.(io.ReadCloser); ok {
			rc.Close()
		}
	})
	br := bufio.NewReader(r)
	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		default:
		}
		rr, _, err := ReadResponses(br) // 这里由于是同步的，会卡住，该如何实现
		if err != nil {
			return err
		}
		for _, r := range rr {
			select {
			case <-ctx.Done():
				return context.Cause(ctx)
			case c.ch <- r:
			}
		}
	}
}

func (c *chatClientResponse) ErrCode() int64 {
	if c.err == nil {
		return 0
	}
	return xerror.ErrCode(c.err, 1)
}

func (c *chatClientResponse) ErrMsg() string {
	if c.err == nil {
		return ""
	}
	return c.err.Error()
}

func (c *chatClientResponse) Unwrap() any {
	return nil
}
