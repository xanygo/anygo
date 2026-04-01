//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-05

package xnaming

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/xanygo/anygo/ds/xbus"
	"github.com/xanygo/anygo/xnet"
)

// Topic 用于 xbus 传递消息的 topic
var Topic = xbus.NewTopic("naming nodes")

type Naming interface {
	Scheme() string
	Lookup(ctx context.Context, idc string, address string) ([]xnet.AddrNode, error)
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

func Lookup(ctx context.Context, scheme string, idc string, address string) ([]xnet.AddrNode, error) {
	n, ok := factories[scheme]
	if !ok {
		return nil, fmt.Errorf("not support such scheme %q", scheme)
	}
	select {
	case <-ctx.Done():
		return nil, context.Cause(ctx)
	default:
	}
	return n.Lookup(ctx, idc, address)
}

func LookupRaw(ctx context.Context, idc string, str string) ([]xnet.AddrNode, error) {
	str = strings.TrimSpace(str)
	if str == "" {
		return nil, nil
	}
	scheme, after, found := strings.Cut(str, "@")
	if !found {
		scheme = ""
		after = str
	}
	return Lookup(ctx, scheme, idc, after)
}

// IsDynamicAddress 是否是需要动态解析的地址列表
func IsDynamicAddress(names ...string) bool {
	for _, name := range names {
		if strings.Contains(name, "@") {
			return true
		}
	}
	return false
}
