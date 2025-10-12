//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-27

package xnet

import (
	"context"
	"encoding"
	"encoding/binary"
	"encoding/json"
	"net"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/internal/zdefine"
	"github.com/xanygo/anygo/xnet/internal"
)

func NewAddr(network, address string) *Addr {
	return &Addr{
		network: network,
		address: address,
		attr:    &xmap.SliceValueSync[string, string]{},
	}
}

var _ net.Addr = (*Addr)(nil)
var _ encoding.TextMarshaler = (*Addr)(nil)
var _ json.Marshaler = (*Addr)(nil)

type Addr struct {
	network string
	address string
	attr    *xmap.SliceValueSync[string, string]
}

func (a *Addr) MarshalJSON() ([]byte, error) {
	data := map[string]any{
		"Network": a.network,
		"Address": a.address,
	}
	return json.Marshal(data)
}

func (a *Addr) MarshalText() ([]byte, error) {
	return a.MarshalJSON()
}

func (a *Addr) Network() string {
	return a.network
}

func (a *Addr) String() string {
	return a.address
}

func (a *Addr) Equal(b net.Addr) bool {
	return a.network == b.Network() && a.address == b.String()
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

var _ zdefine.HasKey[string] = (*AddrNode)(nil)

type AddrNode struct {
	HostPort string
	Addr     net.Addr
}

func (n AddrNode) Key() string {
	return n.Addr.String()
}

func (n AddrNode) Host() string {
	host, _, _ := net.SplitHostPort(n.HostPort)
	return host
}

func (n AddrNode) Port() string {
	_, port, _ := net.SplitHostPort(n.HostPort)
	return port
}

func NewAddrNodes(addrs ...net.Addr) []AddrNode {
	nodes := make([]AddrNode, len(addrs))
	for i, addr := range addrs {
		nodes[i] = AddrNode{
			HostPort: addr.String(),
			Addr:     addr,
		}
	}
	return nodes
}
