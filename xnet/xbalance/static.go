//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-07

package xbalance

import (
	"context"
	"net"
	"sync/atomic"

	"github.com/xanygo/anygo/xnet"
)

var _ Reader = (*Static)(nil)

type Static struct {
	nodes []xnet.AddrNode
	index atomic.Int32
}

func (s *Static) Name() string {
	return NameStatic
}

func (s *Static) Pick(ctx context.Context) (*xnet.AddrNode, error) {
	total := len(s.nodes)
	if total == 0 {
		return nil, ErrEmptyNode
	}
	index := s.index.Add(1) - 1
	idx := int(index % int32(total))
	return &s.nodes[idx], nil
}

func NewStatic(nodes ...xnet.AddrNode) *Static {
	return &Static{nodes: nodes}
}

func NewStaticByAddr(addrs ...net.Addr) *Static {
	return &Static{nodes: xnet.NewAddrNodes(addrs...)}
}
