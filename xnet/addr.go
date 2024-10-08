//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-27

package xnet

import (
	"context"
	"net"

	"github.com/xanygo/anygo/xmap"
)

func NewAddr(network, host string) *Addr {
	return &Addr{
		network: network,
		host:    host,
		attr:    &xmap.SliceValueSync[string, string]{},
	}
}

var _ net.Addr = (*Addr)(nil)

type Addr struct {
	network string
	host    string
	attr    *xmap.SliceValueSync[string, string]
}

func (a *Addr) Network() string {
	return a.network
}

func (a *Addr) String() string {
	return a.host
}

func (a *Addr) Equal(b net.Addr) bool {
	return a.network == b.Network() && a.host == b.String()
}

// Attr 附加属性
func (a *Addr) Attr() *xmap.SliceValueSync[string, string] {
	return a.attr
}

func ContextWithAddr(ctx context.Context, addr net.Addr) context.Context {
	return context.WithValue(ctx, ctxKeyAddr, addr)
}

func AddrFromContext(ctx context.Context) net.Addr {
	addr, _ := ctx.Value(ctxKeyAddr).(net.Addr)
	return addr
}
