//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-30

package xpool

import (
	"context"
	"net"

	"github.com/xanygo/anygo/internal/znet"
)

type (
	ConnPool = Pool[net.Conn]

	ConnGroupPool = GroupPool[net.Addr, net.Conn]
)

var _ Factory[net.Conn] = (*ConnFactory)(nil)
var _ GroupFactory[net.Addr, net.Conn] = (*ConnFactory)(nil)
var _ Validator[net.Conn] = (*ConnFactory)(nil)

type ConnFactory struct {
	Single func(ctx context.Context) (net.Conn, error)
	Group  func(ctx context.Context, addr net.Addr) (net.Conn, error)
}

func (c *ConnFactory) New(ctx context.Context) (net.Conn, error) {
	return c.Single(ctx)
}

func (c *ConnFactory) NewWithKey(ctx context.Context, key net.Addr) (net.Conn, error) {
	if c.Group != nil {
		return c.Group(ctx, key)
	}
	return c.Single(ctx)
}

func (c *ConnFactory) Validate(conn net.Conn) error {
	return znet.ConnCheck(conn)
}
