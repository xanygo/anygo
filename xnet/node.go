//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-12

package xnet

import (
	"crypto/tls"
	"net"
)

type AddrNode struct {
	HostPort string
	Addr     net.Addr
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

type ConnNode struct {
	Conn    net.Conn
	TlsConn *tls.Conn
	Addr    AddrNode
}

func (c *ConnNode) Clone() *ConnNode {
	if c == nil {
		return nil
	}
	return &ConnNode{
		Conn:    c.Conn,
		TlsConn: c.TlsConn,
		Addr:    c.Addr,
	}
}
