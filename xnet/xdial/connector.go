//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-03

package xdial

import (
	"context"
	"time"

	"github.com/xanygo/anygo/ds/xmetric"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xoption"
)

// Connector 网络连接器
type Connector interface {
	Connect(ctx context.Context, addr xnet.AddrNode, opt xoption.Reader) (*xnet.ConnNode, error)
}

var _ Connector = ConnectorFunc(nil)

type ConnectorFunc func(ctx context.Context, addr xnet.AddrNode, opt xoption.Reader) (*xnet.ConnNode, error)

func (c ConnectorFunc) Connect(ctx context.Context, addr xnet.AddrNode, opt xoption.Reader) (*xnet.ConnNode, error) {
	return c(ctx, addr, opt)
}

func Connect(ctx context.Context, c Connector, addr xnet.AddrNode, opt xoption.Reader) (nc *xnet.ConnNode, err error) {
	ctx, span := xmetric.Start(ctx, "Connect")

	defer func() {
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
	// 尝试多次连接，由于在 xnet.DialerImpl 里已经有 Hedging request 的逻辑，所以这里就不需要了
	for i := 0; i < total; i++ {
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
