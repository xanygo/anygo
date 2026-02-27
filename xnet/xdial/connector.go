//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-03

package xdial

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/xanygo/anygo/ds/xctx"
	"github.com/xanygo/anygo/ds/xmetric"
	"github.com/xanygo/anygo/ds/xoption"
	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/xnet"
)

// Connector 网络连接器
type Connector interface {
	Connect(ctx context.Context, addr xnet.AddrNode, opt xoption.Reader) (*xnet.ConnNode, error)
}

type Interceptor struct {
	BeforeConnect func(ctx context.Context, addr xnet.AddrNode, opt xoption.Reader) (context.Context, xnet.AddrNode, xoption.Reader)

	AfterConnect func(ctx context.Context, addr xnet.AddrNode, opt xoption.Reader, result *xnet.ConnNode, err error) (*xnet.ConnNode, error)

	// 在执行 TlsHandshake 前执行
	// 若 *tls.Config == nil， 表名此次连接不需要执行 TlsHandshake
	BeforeTlsHandshake func(ctx context.Context, conn *xnet.ConnNode, opt xoption.Reader, target xnet.AddrNode, tc *tls.Config) (context.Context, *xnet.ConnNode, xoption.Reader, xnet.AddrNode, *tls.Config)

	AfterTlsHandshake func(ctx context.Context, conn *xnet.ConnNode, opt xoption.Reader, target xnet.AddrNode, tc *tls.Config, result *xnet.ConnNode, err error) (*xnet.ConnNode, error)
}

var _ Connector = ConnectorFunc(nil)

type ConnectorFunc func(ctx context.Context, addr xnet.AddrNode, opt xoption.Reader) (*xnet.ConnNode, error)

func (c ConnectorFunc) Connect(ctx context.Context, addr xnet.AddrNode, opt xoption.Reader) (*xnet.ConnNode, error) {
	return c(ctx, addr, opt)
}

func Connect(ctx context.Context, c Connector, addr xnet.AddrNode, opt xoption.Reader) (nc *xnet.ConnNode, err error) {
	ctx, span := xmetric.Start(ctx, "Connect")

	its := allITs(ctx)
	for _, it := range its {
		if it.BeforeConnect == nil {
			continue
		}
		ctx, addr, opt = it.BeforeConnect(ctx, addr, opt)
	}

	defer func() {
		for _, it := range its {
			if it.AfterConnect == nil {
				continue
			}
			nc, err = it.AfterConnect(ctx, addr, opt, nc, err)
		}

		if !span.IsRecording() {
			return
		}
		if conn := nc.NetConn(); conn != nil {
			span.SetAttributes(
				xmetric.AnyAttr("Remote", conn.RemoteAddr()),
				xmetric.AnyAttr("Local", conn.LocalAddr()),
			)
			span.SetAttributes(
				xmetric.AnyAttr("Addr", addr),
			)
		}
		span.RecordError(err)
		span.End()
	}()
	if opt == nil {
		opt = xoption.ReaderFromContext(ctx)
	}

	total := xoption.ConnectRetry(opt) + 1
	timeout := xoption.ConnectTimeout(opt)
	span.SetAttemptCount(total)

	doDial := func(ctx context.Context) (nc *xnet.ConnNode, err error) {
		ctx1, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		if c == nil {
			return DefaultConnector().Connect(ctx1, addr, opt)
		}
		return c.Connect(ctx1, addr, opt)
	}
	// 尝试多次连接，由于在 xnet.HedgingDialer 里已经有 Hedging request 的逻辑，所以这里就不需要了
	for range total {
		nc, err = doDial(ctx)
		if err == nil {
			break
		}
	}
	return nc, err
}

func DefaultConnector() Connector {
	return ConnectorFunc(defaultConnect)
}

func defaultConnect(ctx context.Context, addrNode xnet.AddrNode, opt xoption.Reader) (*xnet.ConnNode, error) {
	addr := addrNode.Addr
	conn, err := xnet.DialContext(ctx, addr.Network(), addr.String())
	node := &xnet.ConnNode{
		Addr:      addrNode,
		Conn:      conn,
		CreatTime: time.Now(),
	}
	return node, err
}

var globalInterceptors []Interceptor

func RegisterIT(its ...Interceptor) {
	globalInterceptors = append(globalInterceptors, its...)
}

var ctxITKey = xctx.NewKey()

func ContextWithIT(ctx context.Context, its ...Interceptor) context.Context {
	return xctx.WithValues(ctx, ctxITKey, its...)
}

func ITFromContext(ctx context.Context) []Interceptor {
	return xctx.Values[*xctx.Key, Interceptor](ctx, ctxITKey, true)
}

func allITs(ctx context.Context) []Interceptor {
	its := ITFromContext(ctx)
	return xslice.SafeMerge(globalInterceptors, its)
}
