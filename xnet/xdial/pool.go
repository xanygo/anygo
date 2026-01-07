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
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xnet"
)

type (
	ConnPool = xpool.Pool[*xnet.ConnNode]

	ConnGroupPool = xpool.GroupPool[xnet.AddrNode, string, *xnet.ConnNode]
)

var _ xpool.Factory[*xnet.ConnNode] = (*ConnFactory)(nil)
var _ xpool.GroupFactory[xnet.AddrNode, string, *xnet.ConnNode] = (*ConnFactory)(nil)
var _ xpool.Validator[*xnet.ConnNode] = (*ConnFactory)(nil)

type ConnFactory struct {
	Addr    xnet.AddrNode
	Connect func(ctx context.Context, addr xnet.AddrNode) (*xnet.ConnNode, error)
}

func (c *ConnFactory) KeyFactory(key xnet.AddrNode) xpool.Factory[*xnet.ConnNode] {
	return &ConnFactory{
		Addr:    key,
		Connect: c.Connect,
	}
}

func (c *ConnFactory) New(ctx context.Context) (*xnet.ConnNode, error) {
	return c.Connect(ctx, c.Addr)
}

func (c *ConnFactory) NewWithKey(ctx context.Context, key xnet.AddrNode) (*xnet.ConnNode, error) {
	return c.Connect(ctx, key)
}

// Validate 验证网络连接是否有效
//
// 第二个参数 err 是 rpc client 交互后返回的 error，可能是底层网络错误，也可能是业务层错误,
func (c *ConnFactory) Validate(conn *xnet.ConnNode, err error) error {
	if err1 := conn.Err(); err1 != nil {
		return err1
	}
	// 以下，判断出是底层网络错误
	if err != nil && xerror.IsClientNetError(err) {
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
//
// 参数 xpool.Option： 连接池配置参数，可为 nil，当为 nil 时，使用默认值：
//
//	{
//	   MaxOpen: 128,
//	   MaxIdle: MaxOpen/4,
//	   MaxLifeTime: 30min,
//	   MaxIdleTime: 10min,
//	   MaxPoolIdleTime: 10 min,
//	}
func NewGroupPool(name string, opt *xpool.Option, cc Connector) (ConnGroupPool, error) {
	opt = opt.Normalization()
	fixOption(opt)
	upperName := strings.ToUpper(name)
	fac, ok := groupPoolFactory[upperName]
	if !ok {
		return nil, fmt.Errorf("cannot create group pool with name %q", name)
	}
	return fac(opt, cc), nil
}

func fixOption(opt *xpool.Option) {
	if opt.MaxOpen <= 0 {
		// 默认最大连接数：128（中等负载服务器）
		// 对于高并发场景，可在业务层显式调高
		opt.MaxOpen = 128
	}

	if opt.MaxIdle <= 0 {
		// 默认空闲数为 1/4 最大连接数，至少 4
		opt.MaxIdle = max(4, opt.MaxOpen/4)
	}
}

func newLong(opt *xpool.Option, cc Connector) ConnGroupPool {
	fac := &ConnFactory{
		Connect: func(ctx context.Context, addr xnet.AddrNode) (*xnet.ConnNode, error) {
			node, err := Connect(ctx, cc, addr, nil)
			if node != nil {
				node.LongPool = true
			}
			return node, err
		},
	}
	return xpool.NewGroupPool[xnet.AddrNode, string, *xnet.ConnNode](opt, fac)
}

func newShort(opt *xpool.Option, cc Connector) ConnGroupPool {
	fac := &ConnFactory{
		Connect: func(ctx context.Context, addr xnet.AddrNode) (*xnet.ConnNode, error) {
			return Connect(ctx, cc, addr, nil)
		},
	}
	return &xpool.DirectGroup[xnet.AddrNode, string, *xnet.ConnNode]{
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
	select {
	case <-ctx1.Done():
		return nil, context.Cause(ctx1)
	default:
	}
	return p.Get(ctx1, addr)
}
