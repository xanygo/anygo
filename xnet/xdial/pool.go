//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-03

package xdial

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/xanygo/anygo/ds/xmeta"
	"github.com/xanygo/anygo/ds/xmetric"
	"github.com/xanygo/anygo/ds/xpool"
	"github.com/xanygo/anygo/internal/znet"
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xnet"
)

// GroupPool service 对象所需要的类型
//
// io.ReadWriteCloser 在大多数情况下是 net.Conn（目前实际为 *xnet.ConnNode）
// 同时要求连接池返回的 ReadWriteCloser 要实现 xmeta.Setter 和 xmeta.Getter
type GroupPool = xpool.GroupPool[xnet.AddrNode, string, io.ReadWriteCloser]

var _ xpool.Factory[io.ReadWriteCloser] = (*Factory)(nil)
var _ xpool.GroupFactory[xnet.AddrNode, string, io.ReadWriteCloser] = (*Factory)(nil)
var _ xpool.Validator[io.ReadWriteCloser] = (*Factory)(nil)

type Factory struct {
	Addr    xnet.AddrNode
	Connect func(ctx context.Context, addr xnet.AddrNode) (io.ReadWriteCloser, error)
}

func (c *Factory) KeyFactory(key xnet.AddrNode) xpool.Factory[io.ReadWriteCloser] {
	return &Factory{
		Addr:    key,
		Connect: c.Connect,
	}
}

func (c *Factory) New(ctx context.Context) (io.ReadWriteCloser, error) {
	return c.Connect(ctx, c.Addr)
}

func (c *Factory) NewWithKey(ctx context.Context, key xnet.AddrNode) (io.ReadWriteCloser, error) {
	return c.Connect(ctx, key)
}

// Validate 验证网络连接是否有效
//
// 第二个参数 err 是 rpc client 交互后返回的 error，可能是底层网络错误，也可能是业务层错误,
func (c *Factory) Validate(rw io.ReadWriteCloser, err error) error {
	if he, ok := rw.(interface{ Err() error }); ok {
		if err1 := he.Err(); err1 != nil {
			return err1
		}
	}

	// 以下，判断出是底层网络错误
	if err != nil && xerror.IsClientNetError(err) {
		return err
	}

	conn, ok := rw.(net.Conn)
	if !ok {
		return err
	}

	uc := xnet.UnwrapConn(conn)
	return znet.ConnCheck(uc)
}

var groupPoolFactory = map[string]func(opt *xpool.Option, cc Connector) GroupPool{}

// RegisterGroupPool 注册创建 GroupPool 的工厂类，注册成功返回 true
func RegisterGroupPool(name string, new func(opt *xpool.Option, cc Connector) GroupPool) bool {
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

// NewGroupPool 使用名称创建 GroupPool，name 支持：Long-长连接，Short-短连接
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
func NewGroupPool(name string, opt *xpool.Option, cc Connector) (GroupPool, error) {
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

func newLong(opt *xpool.Option, cc Connector) GroupPool {
	fac := &Factory{
		Connect: func(ctx context.Context, addr xnet.AddrNode) (io.ReadWriteCloser, error) {
			ctx = xpool.ContextWithOption(ctx, opt)
			node, err := Connect(ctx, cc, addr, nil)
			xmeta.TrySetMeta(node, xmeta.KeyLongPool, true)
			return node, err
		},
	}
	return xpool.NewGroupPool[xnet.AddrNode, string, io.ReadWriteCloser](opt, fac)
}

func newShort(opt *xpool.Option, cc Connector) GroupPool {
	fac := &Factory{
		Connect: func(ctx context.Context, addr xnet.AddrNode) (io.ReadWriteCloser, error) {
			ctx = xpool.ContextWithOption(ctx, opt)
			return Connect(ctx, cc, addr, nil)
		},
	}
	return &xpool.DirectGroup[xnet.AddrNode, string, io.ReadWriteCloser]{
		Factory: fac,
	}
}

func GroupPoolGet(ctx context.Context, p GroupPool, addr xnet.AddrNode) (entry xpool.Entry[io.ReadWriteCloser], err error) {
	ctx1, span := xmetric.Start(ctx, "ConnPoolGet")
	if !span.IsRecording() {
		return p.Get(ctx, addr)
	}
	defer func() {
		if err == nil && entry != nil {
			if conn, ok := entry.Raw().(net.Conn); ok && conn != nil {
				span.SetAttributes(
					xmetric.AnyAttr("Remote", conn.RemoteAddr()),
					xmetric.AnyAttr("Local", conn.LocalAddr()),
				)
			}
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
