//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-03

package xdial

import (
	"context"
	"fmt"
	"strings"

	"github.com/xanygo/anygo/ds/xmetric"
	"github.com/xanygo/anygo/ds/xpool"
	"github.com/xanygo/anygo/internal/znet"
	"github.com/xanygo/anygo/xnet"
)

type (
	ConnPool = xpool.Pool[*xnet.ConnNode]

	ConnGroupPool = xpool.GroupPool[xnet.AddrNode, *xnet.ConnNode]
)

var _ xpool.Factory[*xnet.ConnNode] = (*ConnFactory)(nil)
var _ xpool.GroupFactory[xnet.AddrNode, *xnet.ConnNode] = (*ConnFactory)(nil)
var _ xpool.Validator[*xnet.ConnNode] = (*ConnFactory)(nil)

type ConnFactory struct {
	Single func(ctx context.Context) (*xnet.ConnNode, error)
	Group  func(ctx context.Context, addr xnet.AddrNode) (*xnet.ConnNode, error)
}

func (c *ConnFactory) New(ctx context.Context) (*xnet.ConnNode, error) {
	return c.Single(ctx)
}

func (c *ConnFactory) NewWithKey(ctx context.Context, key xnet.AddrNode) (*xnet.ConnNode, error) {
	if c.Group != nil {
		return c.Group(ctx, key)
	}
	return c.Single(ctx)
}

func (c *ConnFactory) Validate(conn *xnet.ConnNode) error {
	if err := conn.Err(); err != nil {
		return err
	}
	uc := xnet.UnwrapConn(conn)
	return znet.ConnCheck(uc)
}

var groupPoolFactory = map[string]func(opt *xpool.Option, cc Connector) ConnGroupPool{}

// RegisterGroupPool 注册创建 ConnGroupPool 的工厂类，注册成功返回 true
func RegisterGroupPool(name string, new func(opt *xpool.Option, cc Connector) ConnGroupPool) bool {
	upperName := strings.ToUpper(name)
	if _, ok := groupPoolFactory[upperName]; ok {
		return false
	}
	groupPoolFactory[upperName] = new
	return true
}

const (
	Long  = "Long"
	Short = "Short"
)

func init() {
	RegisterGroupPool(Long, newLong)
	RegisterGroupPool(Short, newShort)
}

// NewGroupPool 使用名称创建 ConnGroupPool，name 支持：Long-长连接，Short-短连接
func NewGroupPool(name string, opt *xpool.Option, cc Connector) (ConnGroupPool, error) {
	upperName := strings.ToUpper(name)
	fac, ok := groupPoolFactory[upperName]
	if !ok {
		return nil, fmt.Errorf("cannot create group pool with name %q", name)
	}
	return fac(opt, cc), nil
}

func newLong(opt *xpool.Option, cc Connector) ConnGroupPool {
	fac := &ConnFactory{
		Group: func(ctx context.Context, addr xnet.AddrNode) (*xnet.ConnNode, error) {
			return Connect(ctx, cc, addr, nil)
		},
	}
	return xpool.NewGroupPool[xnet.AddrNode, *xnet.ConnNode](opt, fac)
}

func newShort(opt *xpool.Option, cc Connector) ConnGroupPool {
	fac := &ConnFactory{
		Group: func(ctx context.Context, addr xnet.AddrNode) (*xnet.ConnNode, error) {
			return Connect(ctx, cc, addr, nil)
		},
	}
	return &xpool.DirectGroup[xnet.AddrNode, *xnet.ConnNode]{
		Factory: fac,
	}
}

func GroupPoolGet(ctx context.Context, p ConnGroupPool, addr xnet.AddrNode) (entry xpool.Entry[*xnet.ConnNode], err error) {
	ctx1, span := xmetric.Start(ctx, "ConnPoolGet")
	if !span.IsRecording() {
		return p.Get(ctx, addr)
	}
	defer func() {
		if err == nil && entry != nil {
			conn := entry.Object()
			nc := conn.Outer()
			span.SetAttributes(
				xmetric.AnyAttr("Remote", nc.RemoteAddr()),
				xmetric.AnyAttr("Local", nc.LocalAddr()),
			)
		}
		span.RecordError(err)
		span.End()
	}()
	return p.Get(ctx1, addr)
}
