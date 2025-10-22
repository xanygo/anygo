//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-05

package xnaming

import (
	"crypto/md5"
	"encoding/hex"
	"sort"

	"github.com/xanygo/anygo/ds/xbus"
	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/xnet"
)

func newNodeProducer() *nodeProducer {
	return &nodeProducer{
		ch: make(chan xbus.Message, 1),
	}
}

var _ xbus.Producer = (*nodeProducer)(nil)

type nodeProducer struct {
	ch    chan xbus.Message
	nodes xsync.Value[[]xnet.AddrNode]
	sign  xsync.Value[string]
}

func (nw *nodeProducer) Messages() <-chan xbus.Message {
	return nw.ch
}

func (nw *nodeProducer) Nodes() []xnet.AddrNode {
	return nw.nodes.Load()
}

func (nw *nodeProducer) Update(nodes []xnet.AddrNode) {
	var newSign string
	if len(nodes) == len(nw.nodes.Load()) {
		newSign = nw.genSign(nodes)
		if nw.sign.Load() == newSign {
			return
		}
	}
	nw.nodes.Store(nodes)
	nw.sign.Store(newSign)
	select {
	case nw.ch <- xbus.Message{
		Topic:   Topic,
		Payload: nodes,
	}:
	default:
	}
}

func (nw *nodeProducer) genSign(nodes []xnet.AddrNode) string {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Addr.String() < nodes[j].Addr.String()
	})
	h := md5.New()
	for _, node := range nodes {
		_, _ = h.Write([]byte(node.Addr.String()))
	}
	return hex.EncodeToString(h.Sum(nil))
}
