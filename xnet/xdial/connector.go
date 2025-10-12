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

	timeout := xoption.ConnectTimeout(opt)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if c == nil {
		return DefaultConnector().Connect(ctx, addr, opt)
	}
	return c.Connect(ctx, addr, opt)
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
