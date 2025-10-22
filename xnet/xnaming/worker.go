//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-05

package xnaming

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/xanygo/anygo/ds/xbus"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xpp"
)

func NewWorker(idc string, primary []string, fallback []string) (*Worker, error) {
	if len(primary) == 0 && len(fallback) == 0 {
		return nil, errors.New("no items")
	}
	return &Worker{
		itemsPrimary:  primary,
		itemsFallback: fallback,
		idc:           idc,
	}, nil
}

var _ xpp.CycleWorker = (*Worker)(nil)
var _ xbus.Producer = (*Worker)(nil)

type Worker struct {
	idc           string
	itemsPrimary  []string
	itemsFallback []string
	worker        *xpp.CycleWorkerTpl
	once          sync.Once
	producer      *nodeProducer
}

func (n *Worker) Name() string {
	return "CycleWorker"
}

func (n *Worker) Nodes() []xnet.AddrNode {
	n.once.Do(n.initOnce)
	return n.producer.Nodes()
}

func (n *Worker) initOnce() {
	n.worker = &xpp.CycleWorkerTpl{
		Do: n.do,
	}
	n.producer = newNodeProducer()
}

func (n *Worker) Start(ctx context.Context, cycle time.Duration) error {
	n.once.Do(n.initOnce)
	return n.worker.Start(ctx, cycle)
}

func (n *Worker) Messages() <-chan xbus.Message {
	n.once.Do(n.initOnce)
	return n.producer.Messages()
}

func (n *Worker) do(ctx context.Context) error {
	primaryNodes, err1 := n.search(ctx, n.idc, n.itemsPrimary)
	if len(primaryNodes) > 0 {
		n.producer.Update(primaryNodes)
		return nil
	}

	fallbackNodes, err2 := n.search(ctx, n.idc, n.itemsFallback)
	if len(fallbackNodes) > 0 {
		n.producer.Update(fallbackNodes)
		return nil
	}
	if err1 != nil {
		return err1
	}
	return err2
}

func (n *Worker) search(ctx context.Context, idc string, items []string) ([]xnet.AddrNode, error) {
	var allNodes []xnet.AddrNode
	var errs []error
	for idx, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		scheme, after, found := strings.Cut(item, "://")
		if !found {
			scheme = ""
			after = item
		}
		name, paramStr, _ := strings.Cut(after, "|")

		name = strings.TrimSpace(name)
		paramStr = strings.TrimSpace(paramStr)

		var param url.Values
		if len(paramStr) > 0 {
			var err error
			param, err = url.ParseQuery(paramStr)
			if err != nil {
				errs = append(errs, fmt.Errorf("[%d]=%q %worker", idx, item, err))
				continue
			}
		}
		nodes, err := Lookup(ctx, scheme, idc, name, param)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		allNodes = append(allNodes, nodes...)
	}
	if len(errs) == 0 {
		if len(allNodes) > 0 {
			return allNodes, nil
		}
		return nil, errors.New("no valid nodes")
	}
	return allNodes, errors.Join(errs...)
}

func (n *Worker) Stop(ctx context.Context) error {
	n.once.Do(n.initOnce)

	return n.worker.Stop(ctx)
}
