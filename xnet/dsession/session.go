//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-04-02

package dsession

import (
	"context"
	"io"
	"time"

	"github.com/xanygo/anygo/ds/xctx"
	"github.com/xanygo/anygo/ds/xmeta"
	"github.com/xanygo/anygo/ds/xmetric"
	"github.com/xanygo/anygo/ds/xoption"
	"github.com/xanygo/anygo/xio"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xdial"
)

// Reply 表示 Starter 执行后的结果
type Reply interface {
	String() string // 用于打印出完整的内容

	// Summary 返回简短的描述信息
	Summary() string
}

// Starter 用于创建连接后,tcp client 业务层面的握手
// 如 Redis 协议发送 Hello Request
type Starter interface {
	// StartSession 开始一个会话，conn 大多数情况下是 *xnet.ConnNode
	StartSession(ctx context.Context, rw io.ReadWriter, opt xoption.Reader) (Reply, error)
}

var _ Starter = StartFunc(nil)

type StartFunc func(ctx context.Context, rw io.ReadWriter, opt xoption.Reader) (Reply, error)

func (h StartFunc) StartSession(ctx context.Context, rw io.ReadWriter, opt xoption.Reader) (Reply, error) {
	return h(ctx, rw, opt)
}

func WithStarter(c xdial.Connector, h Starter, opt xoption.Reader) xdial.Connector {
	return xdial.ConnectorFunc(func(ctx context.Context, addr xnet.AddrNode, opt xoption.Reader) (io.ReadWriteCloser, error) {
		conn, err := c.Connect(ctx, addr, opt)
		if err != nil {
			return conn, err
		}
		ret, err := h.StartSession(ctx, conn, opt)
		if err != nil {
			conn.Close()
			return conn, err
		}
		xmeta.TrySetMeta(conn, xmeta.KeySessionReply, ret)
		return conn, nil
	})
}

// StartSession 用于在连接创建完成后，业务正式使用前，执行会话开启的逻辑。
// 如身份认证，协议升级等。
// 目前在 xservice.connector 里调用
func StartSession(ctx context.Context, rw io.ReadWriter, opt xoption.Reader) (ret Reply, started bool, err error) {
	ctx, span := xmetric.Start(ctx, "StartSession")
	timeout := xoption.HandshakeTimeout(opt)
	defer func() {
		span.SetAttributes(
			xmetric.AnyAttr("Timeout", timeout),
		)
		span.RecordError(err)
		span.End()
	}()

	handler := FromContext(ctx)
	if handler == nil {
		if cfg := xoption.SessionStarter(opt); cfg != nil {
			handler, err = create(cfg)
			if err != nil {
				return nil, false, err
			}
		}
	}
	if handler == nil {
		protocol := xoption.Protocol(opt)
		span.SetAttributes(xmetric.AnyAttr("Protocol", protocol))
		handler = FindProtocol(protocol)
	}
	// 若找不到，则直接返回(跳过)
	if handler == nil {
		return nil, false, nil
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if ds, ok := rw.(xio.DeadlineSetter); ok {
		ds.SetDeadline(time.Now().Add(timeout))
		defer ds.SetDeadline(time.Time{})
	}
	ret, err = handler.StartSession(ctx, rw, opt)

	if err == nil && ret != nil {
		span.SetAttributes(xmetric.AnyAttr("Summary", ret.Summary()))
	}
	return ret, true, err
}

var ctxKey = xctx.NewKey()

func ContextWith(ctx context.Context, h Starter) context.Context {
	return context.WithValue(ctx, ctxKey, h)
}

func FromContext(ctx context.Context) Starter {
	val, _ := ctx.Value(ctxKey).(Starter)
	return val
}
