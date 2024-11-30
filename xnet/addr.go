//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-27

package xnet

import (
	"context"
	"encoding/binary"
	"net"

	"github.com/xanygo/anygo/xmap"
	"github.com/xanygo/anygo/xnet/internal"
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

func IP4ToLong(ip net.IP) uint32 {
	parsedIP := ip.To4()
	if parsedIP == nil {
		return 0
	}
	return binary.BigEndian.Uint32(parsedIP)
}

func LongToIP4(long uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, long)
	return ip
}

// IsIPAddress  判断传入的 host 是否是一个 ip
func IsIPAddress(host string) bool {
	ip, _ := internal.ParseIPZone(host)
	return ip != nil
}
