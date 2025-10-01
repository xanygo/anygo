//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-30

package xredis

import (
	"context"
	"io"
	"time"

	"github.com/xanygo/anygo/ds/xpool"
	"github.com/xanygo/anygo/store/xredis/resp3"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xbalance"
)

type ConnPool = xpool.Pool[*Conn]

var _ io.Closer = (*Conn)(nil)

type Conn struct {
	Conn *resp3.Conn
}

func (c *Conn) Close() error {
	return c.Conn.Close()
}

var _ xpool.Factory[*Conn] = (*connFactory)(nil)

type connFactory struct {
	AP    xbalance.Reader
	Hello resp3.HelloRequest
}

func (c *connFactory) New(ctx context.Context) (*Conn, error) {
	addr, err := c.AP.Pick(ctx)
	if err != nil {
		return nil, err
	}
	conn, err := xnet.DialContext(ctx, addr.Addr.Network(), addr.Addr.String())
	if err != nil {
		return nil, err
	}
	cc := resp3.NewConn(conn, 3*time.Second)
	resp, err := cc.Send(c.Hello)
	if err != nil {
		conn.Close()
		return nil, err
	}
	_ = resp // todo
	return &Conn{Conn: cc}, nil
}
