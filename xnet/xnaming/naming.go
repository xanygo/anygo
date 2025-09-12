//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-05

package xnaming

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/xanygo/anygo/xbus"
	"github.com/xanygo/anygo/xnet"
)

// Topic 用于 xbus 传递消息的 topic
var Topic = xbus.NewTopic("naming nodes")

type Naming interface {
	Scheme() string
	Lookup(ctx context.Context, idc string, name string, param url.Values) ([]xnet.AddrNode, error)
}

var factories = map[string]Naming{}

func Register(n Naming) error {
	if _, ok := factories[n.Scheme()]; ok {
		return errors.New("duplicated scheme " + n.Scheme())
	}
	factories[n.Scheme()] = n
	return nil
}

func MustRegister(n Naming) {
	if err := Register(n); err != nil {
		panic(err)
	}
}

func Lookup(ctx context.Context, scheme string, idc string, name string, param url.Values) ([]xnet.AddrNode, error) {
	n, ok := factories[scheme]
	if !ok {
		return nil, fmt.Errorf("not support such scheme %q", scheme)
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return n.Lookup(ctx, idc, name, param)
}
