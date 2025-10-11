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
	"time"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/store/xredis/resp3"
	"github.com/xanygo/anygo/xattr"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xdial"
	"github.com/xanygo/anygo/xnet/xrpc"
	"github.com/xanygo/anygo/xnet/xservice"
	"github.com/xanygo/anygo/xoption"
	"github.com/xanygo/anygo/xpp"
)

type (
	Request = resp3.Request

	Result = resp3.Result
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

func (c *Client) do(ctx context.Context, cmd Request) *rpcResponse {
	req := &rpcRequest{
		req: cmd,
	}
	resp := &rpcResponse{}
	_ = xrpc.Invoke(ctx, c.Service, req, resp, c.once.Load()...)
	return resp
}

func (c *Client) Do(ctx context.Context, cmd Request) Response {
	resp := c.do(ctx, cmd)
	return Response{
		result: resp.result,
		err:    resp.err,
	}
}

func init() {
	handler := xdial.HandshakeFunc(handshake)
	xdial.RegisterHandshakeHandler(Protocol, handler)
	xdial.RegisterHandshakeHandler("Redis", handler)
}

// 创建连接后，和 redis server 握手
func handshake(ctx context.Context, conn *xnet.ConnNode, opt xoption.Reader) (any, error) {
	hello := resp3.HelloRequest{}
	cfg := xoption.Extra(opt, "Redis")
	var dbIndex int
	if cfg != nil {
		mp, ok := cfg.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("config [Redis] part is not map[string]any, %#v", cfg)
		}
		if username, ok1 := xmap.GetString(mp, fieldUsername); ok1 {
			hello.Username = username
		}
		if password, ok1 := xmap.GetString(mp, fieldPassword); ok1 {
			hello.Password = password
		}
		if num, ok1 := xmap.GetInt64(mp, fieldDBIndex); ok1 && num > 0 {
			dbIndex = int(num)
		}
	}
	cc := conn.NetConn()
	bf := bp.Get()
	_, err := cc.Write(hello.Bytes(bf))
	bp.Put(bf)
	if err != nil {
		return nil, err
	}
	br := bufio.NewReader(cc)
	result, err := resp3.ReadByType(br, hello.ResponseType())
	if err != nil || dbIndex == 0 {
		return result, err
	}

	// --------------------------------------------------------------
	// 发送 SELECT 命令选择数据库
	sc := hello.Select()
	bf = bp.Get()
	_, err = cc.Write(sc.Bytes(bf))
	bp.Put(bf)
	if err != nil {
		return result, err
	}
	result2, err2 := resp3.ReadByType(br, sc.ResponseType())
	err = resp3.ToOkStatus(result2, err2)
	return result, err
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
	dbIndex, err := strconv.Atoi(strings.TrimLeft(uu.Path, "/"))
	if err != nil || dbIndex < 0 {
		return nil, nil, fmt.Errorf("invalid db index: %s", uu.Path)
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
		err = xpp.TryStartWorker(context.Background(), 10*time.Minute, ser)
	}
	if err != nil {
		return nil, nil, err
	}

	return ser, NewClient(ser), nil
}
