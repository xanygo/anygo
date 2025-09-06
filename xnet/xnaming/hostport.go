//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-05

package xnaming

import (
	"context"
	"net"
	"net/url"

	"github.com/xanygo/anygo/xnet"
)

var _ Naming = (*HostPort)(nil)

type HostPort struct{}

func (i *HostPort) Scheme() string {
	return ""
}

func (i *HostPort) Lookup(ctx context.Context, idc string, hostPort string, param url.Values) ([]Node, error) {
	_, _, err := net.SplitHostPort(hostPort)
	if err != nil {
		return nil, err
	}
	addr := xnet.NewAddr("tcp", hostPort)
	node := NewNode(hostPort, addr)
	return []Node{node}, nil
}

func init() {
	MustRegister(&HostPort{})
}
