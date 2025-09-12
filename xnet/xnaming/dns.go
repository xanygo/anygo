//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-06

package xnaming

import (
	"context"
	"net"
	"net/url"

	"github.com/xanygo/anygo/xnet"
)

var _ Naming = (*DNS)(nil)

type DNS struct {
}

func (d *DNS) Scheme() string {
	return "dns"
}

func (d *DNS) Lookup(ctx context.Context, idc string, hostPort string, param url.Values) ([]xnet.AddrNode, error) {
	host, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		return nil, err
	}
	ips, err := xnet.LookupIP(ctx, "tcp", host)
	if err != nil {
		return nil, err
	}
	nodes := make([]xnet.AddrNode, 0, len(ips))
	for _, ip := range ips {
		addr := net.JoinHostPort(ip.String(), port)
		node := xnet.AddrNode{
			HostPort: hostPort,
			Addr:     xnet.NewAddr("tcp", addr),
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func init() {
	MustRegister(&DNS{})
}
