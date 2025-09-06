//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-05

package xnaming

import (
	"crypto/md5"
	"encoding/hex"
	"net"
	"sort"

	"github.com/xanygo/anygo/xbus"
	"github.com/xanygo/anygo/xsync"
)

type Node interface {
	// Name 原始名称，如当传入 一个域名的时候，此值为域名，若是 HostPort 则是 HostPort
	Name() string

	// Addr 地址
	Addr() net.Addr
}

func NewNode(name string, addr net.Addr) Node {
	return &nodeImpl{
		name: name,
		addr: addr,
	}
}

var _ Node = (*nodeImpl)(nil)

type nodeImpl struct {
	name string
	addr net.Addr
}

func (n *nodeImpl) Name() string {
	return n.name
}

func (n *nodeImpl) Addr() net.Addr {
	return n.addr
}

func newNodeProducer() *nodeProducer {
	return &nodeProducer{
		ch: make(chan xbus.Message, 1),
	}
}

var _ xbus.Producer = (*nodeProducer)(nil)

type nodeProducer struct {
	ch    chan xbus.Message
	nodes xsync.Value[[]Node]
	sign  xsync.Value[string]
}

func (nw *nodeProducer) Messages() <-chan xbus.Message {
	return nw.ch
}

func (nw *nodeProducer) Nodes() []Node {
	return nw.nodes.Load()
}

func (nw *nodeProducer) Update(nodes []Node) {
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

func (nw *nodeProducer) genSign(nodes []Node) string {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Addr().String() < nodes[j].Addr().String()
	})
	h := md5.New()
	for _, node := range nodes {
		_, _ = h.Write([]byte(node.Addr().String()))
	}
	return hex.EncodeToString(h.Sum(nil))
}
