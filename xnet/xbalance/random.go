//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-05

package xbalance

import (
	"context"
	"math/rand/v2"
	"sync"

	"github.com/xanygo/anygo/xnet/xnaming"
)

var _ LoadBalancer = (*Random)(nil)

type Random struct {
	nodes []xnaming.Node
	rw    sync.RWMutex
}

func (r *Random) Name() string {
	return NameRandom
}

func (r *Random) Pick(_ context.Context) (xnaming.Node, error) {
	r.rw.RLock()
	defer r.rw.RUnlock()
	total := len(r.nodes)
	if total == 0 {
		return nil, ErrEmptyNode
	}
	idx := rand.IntN(total)
	return r.nodes[idx], nil
}

func (r *Random) Init(param any, nodes []xnaming.Node) error {
	return r.Update(context.Background(), nodes)
}

func (r *Random) Update(ctx context.Context, nodes []xnaming.Node) error {
	r.rw.Lock()
	r.nodes = nodes
	r.rw.Unlock()
	return nil
}
