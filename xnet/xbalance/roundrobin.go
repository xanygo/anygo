//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-05

package xbalance

import (
	"context"
	"sync"

	"github.com/xanygo/anygo/xnet"
)

var _ LoadBalancer = (*RoundRobin)(nil)

type RoundRobin struct {
	nodes []xnet.AddrNode
	rw    sync.Mutex
	index int64
}

func (r *RoundRobin) Name() string {
	return NameRoundRobin
}

func (r *RoundRobin) Pick(ctx context.Context) (*xnet.AddrNode, error) {
	r.rw.Lock()
	defer r.rw.Unlock()
	total := len(r.nodes)
	if total == 0 {
		return nil, ErrEmptyNode
	}
	r.index++
	idx := int(r.index % int64(total))
	return &r.nodes[idx], nil
}

func (r *RoundRobin) Init(param any, nodes []xnet.AddrNode) error {
	return r.Update(context.Background(), nodes)
}

func (r *RoundRobin) Update(ctx context.Context, nodes []xnet.AddrNode) error {
	r.rw.Lock()
	r.nodes = nodes
	r.rw.Unlock()
	return nil
}
