//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-03

package xdial

import (
	"context"

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

func Connect(ctx context.Context, c Connector, addr xnet.AddrNode, opt xoption.Reader) (*xnet.ConnNode, error) {
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
	conn, err := xnet.DefaultDialer.DialContext(ctx, addr.Network(), addr.String())
	node := &xnet.ConnNode{
		Addr: addrNode,
		Conn: conn,
	}
	return node, err
}
