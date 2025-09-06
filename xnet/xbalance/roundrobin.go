//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-05

package xbalance

import (
	"context"
	"sync"

	"github.com/xanygo/anygo/xnet/xnaming"
)

var _ LoadBalancer = (*RoundRobin)(nil)

type RoundRobin struct {
	nodes []xnaming.Node
	rw    sync.RWMutex
	index int64
}

func (r *RoundRobin) Name() string {
	return NameRoundRobin
}

func (r *RoundRobin) Pick(ctx context.Context) (xnaming.Node, error) {
	r.rw.RLock()
	defer r.rw.RUnlock()
	total := len(r.nodes)
	if total == 0 {
		return nil, ErrEmptyNode
	}
	r.index++
	idx := int(r.index % int64(total))
	return r.nodes[idx], nil
}

func (r *RoundRobin) Init(param any, nodes []xnaming.Node) error {
	return r.Update(context.Background(), nodes)
}

func (r *RoundRobin) Update(ctx context.Context, nodes []xnaming.Node) error {
	r.rw.Lock()
	defer r.rw.Unlock()
	r.nodes = nodes
	return nil
}
