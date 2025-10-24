//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-30

package xredis

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/xanygo/anygo/ds/xcast"
	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/ds/xoption"
	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/store/xredis/resp3"
	"github.com/xanygo/anygo/xattr"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xdial"
	"github.com/xanygo/anygo/xnet/xrpc"
	"github.com/xanygo/anygo/xnet/xservice"
	"github.com/xanygo/anygo/xpp"
)

const Protocol = "RESP3"

var ErrNil = resp3.ErrNil

func NewClient(service any) *Client {
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
	options := make([]xrpc.Option, 0, 1)
	if c.Registry != nil {
		options = append(options, xrpc.OptServiceRegistry(c.Registry))
	}
	return options
}

func (c *Client) do(ctx context.Context, cmd resp3.Request) *rpcResponse {
	req := &rpcRequest{
		req: cmd,
	}
	resp := &rpcResponse{}
	err := c.invoke(ctx, req, resp)
	if resp.err == nil {
		resp.err = err
	}
	return resp
}

func (c *Client) invoke(ctx context.Context, req xrpc.Request, resp xrpc.Response) error {
	return xrpc.Invoke(ctx, c.Service, req, resp, c.once.Load()...)
}

// Do 执行任意的命令，若是调用 Client 没有的方法，可以使用此方法
func (c *Client) Do(ctx context.Context, cmd Cmder) error {
	req := resp3.NewRequest(resp3.DataTypeAny, cmd.Args()...)
	resp := c.do(ctx, req)
	cmd.SetReply(resp.result, resp.err)
	return cmd.Err()
}

func init() {
	handler := xdial.HandshakeFunc(handshake)
	xdial.RegisterHandshakeHandler(Protocol, handler)
	xdial.RegisterHandshakeHandler("Redis", handler)
}

// 创建连接后，和 redis server 握手
//
// xrpc 里有统一的 handshake timeout 设置
// https://redis.io/docs/latest/commands/hello/
func handshake(ctx context.Context, conn *xnet.ConnNode, opt xoption.Reader) (xdial.HandshakeReply, error) {
	hello := resp3.HelloRequest{}
	const redisKey = "Redis"
	cfg := xoption.Extra(opt, redisKey)
	var dbIndex int
	var err error
	xmap.Range[string, any](cfg, func(key string, val any) bool {
		var ok bool
		switch key {
		case fieldUsername:
			hello.Username, ok = xcast.String(val)
		case fieldPassword:
			hello.Password, ok = xcast.String(val)
		case fieldDBIndex:
			dbIndex, ok = xcast.Integer[int](val)
		default:
			ok = true
		}
		if !ok {
			err = fmt.Errorf("invalid filed %s.%s=%#v", redisKey, key, val)
		}
		return ok
	})
	if err != nil {
		return nil, err
	}

	bf := bp.Get()
	_, err = conn.Write(hello.Bytes(bf))
	bp.Put(bf)
	if err != nil {
		return nil, err
	}
	br := bufio.NewReader(conn)
	result, err := resp3.ReadByType(br, hello.ResponseType())
	var mp resp3.Map
	if err == nil {
		var ok bool
		mp, ok = result.(resp3.Map)
		if !ok {
			err = fmt.Errorf("response not map %#v", result)
		}
	}
	md, err := resp3.ToAnyMap(mp, err)
	if err != nil {
		return nil, err
	}

	hp := &resp3.HelloResponse{}
	err = hp.FromMap(md)
	if err != nil {
		return nil, err
	}
	if dbIndex == 0 {
		return hp, nil
	}

	// --------------------------------------------------------------
	// 发送 SELECT 命令选择数据库
	sc := hello.Select()
	bf = bp.Get()
	_, err = conn.Write(sc.Bytes(bf))
	bp.Put(bf)
	if err != nil {
		return hp, err
	}
	result2, err2 := resp3.ReadByType(br, sc.ResponseType())
	err = resp3.ToOkStatus(result2, err2)
	return hp, err
}

const (
	fieldUsername = "Username"
	fieldPassword = "Password"
	fieldDBIndex  = "DBIndex"
)

// NewClientByURI 使用 uri 创建一个 client
// Server URI on format redis://user:password@host:port/dbnum
//
//	User, password and dbnum are optional. For authentication
//	without a username, use username 'default'. For TLS, use
//	the scheme 'rediss'
func NewClientByURI(name string, uri string) (xservice.Service, *Client, error) {
	uu, err := url.Parse(uri)
	if err != nil {
		return nil, nil, err
	}
	switch uu.Scheme {
	case "redis", "rediss":
	default:
		return nil, nil, fmt.Errorf("invalid redis uri: %s", uri)
	}
	if uu.Hostname() == "" || uu.Port() == "" {
		return nil, nil, fmt.Errorf("invalid redis uri %s", uri)
	}
	cfg := &xservice.Config{
		Name:     name,
		Protocol: Protocol,
		ConnPool: &xservice.ConnPoolPart{
			Name: xdial.Long,
		},
		DownStream: xservice.DownStreamPart{
			Address: []string{uu.Host},
		},
	}
	if uu.Scheme == "rediss" {
		cfg.TLS = &xoption.TLSConfig{
			ServerName: uu.Hostname(),
		}
	}
	var dbIndex int
	if uu.Path != "" {
		dbIndex, err = strconv.Atoi(strings.TrimLeft(uu.Path, "/"))
		if err != nil || dbIndex < 0 {
			return nil, nil, fmt.Errorf("invalid db index: %s", uu.Path)
		}
	}

	psw, _ := uu.User.Password()
	cfg.Extra = map[string]any{
		"Redis": map[string]any{
			fieldUsername: uu.User.Username(),
			fieldPassword: psw,
			fieldDBIndex:  dbIndex,
		},
	}
	ser, err := cfg.Parser(xattr.IDC())
	if err == nil {
		err = xpp.TryStartWorker(context.Background(), ser)
	}
	if err != nil {
		return nil, nil, err
	}

	return ser, NewClient(ser), nil
}
