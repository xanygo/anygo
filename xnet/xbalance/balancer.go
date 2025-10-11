//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-05

package xbalance

import (
	"context"
	"errors"
	"fmt"

	"github.com/xanygo/anygo/ds/xmetric"
	"github.com/xanygo/anygo/safely"
	"github.com/xanygo/anygo/xbus"
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xnaming"
	"github.com/xanygo/anygo/xoption"
)

var ErrEmptyNode = errors.New("empty node")

const (
	NameRoundRobin = "RoundRobin"
	NameRandom     = "Random"
	NameStatic     = "Static"
)

type (
	LoadBalancer interface {
		Reader
		Writer
	}

	Reader interface {
		Name() string

		// Pick 选择一个节点，用于发送请求
		Pick(ctx context.Context) (*xnet.AddrNode, error)
	}

	Writer interface {
		// Init 首次初始化
		Init(param any, nodes []xnet.AddrNode) error

		// Update 更新节点列表（动态服务发现时用）
		Update(ctx context.Context, nodes []xnet.AddrNode) error
	}
)

var factories = map[string]func() LoadBalancer{
	NameRandom: func() LoadBalancer {
		return &Random{}
	},
	"rr": func() LoadBalancer {
		return &RoundRobin{}
	},
	NameRoundRobin: func() LoadBalancer {
		return &RoundRobin{}
	},
}

func Register(factory func() LoadBalancer) error {
	if factory == nil {
		return errors.New("factory is nil")
	}
	var name string
	err := safely.Run(func() {
		name = factory().Name()
	})
	if err != nil {
		return err
	}
	if _, has := factories[name]; has {
		return fmt.Errorf("balancer %q already registered", name)
	}
	factories[name] = factory
	return nil
}

func New(name string) (LoadBalancer, error) {
	if name == "" {
		name = NameRoundRobin
	}
	factory, has := factories[name]
	if !has {
		return nil, fmt.Errorf("balancer %q %w", name, xerror.NotFound)
	}
	lb := factory()
	return &worker{
		b: lb,
	}, nil
}

var _ LoadBalancer = (*worker)(nil)
var _ xbus.Consumer = (*worker)(nil)

type worker struct {
	b LoadBalancer
}

func (w *worker) Name() string {
	return w.b.Name()
}

func (w *worker) Pick(ctx context.Context) (*xnet.AddrNode, error) {
	if ap := ReaderFromContext(ctx); ap != nil {
		return ap.Pick(ctx)
	}
	if opt := xoption.ReaderFromContext(ctx); opt != nil {
		if ap := OptReader(opt); ap != nil {
			return ap.Pick(ctx)
		}
	}
	return w.b.Pick(ctx)
}

func (w *worker) Init(param any, nodes []xnet.AddrNode) error {
	return w.b.Init(param, nodes)
}

func (w *worker) Update(ctx context.Context, nodes []xnet.AddrNode) error {
	return w.b.Update(ctx, nodes)
}

func (w *worker) Consume(ctx context.Context, msg xbus.Message) error {
	if msg.Topic != xnaming.Topic {
		return nil
	}
	nodes, ok := msg.Payload.([]xnet.AddrNode)
	if !ok {
		return fmt.Errorf("invalid payload: %T", msg.Payload)
	}
	return w.Update(ctx, nodes)
}

func Pick(ctx context.Context, ap Reader) (addr *xnet.AddrNode, err error) {
	ctx1, span := xmetric.Start(ctx, "Pick")
	defer func() {
		if addr != nil {
			span.SetAttributes(
				xmetric.AnyAttr("Addr", addr),
			)
		}
		span.RecordError(err)
		span.End()
	}()
	return ap.Pick(ctx1)
}
