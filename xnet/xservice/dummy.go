//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-12

package xservice

import (
	"context"
	"time"

	"github.com/xanygo/anygo"
	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/ds/xtype"
	"github.com/xanygo/anygo/xnet"
)

const Dummy = xnet.Dummy

// GetDummyService 从默认的 Registry 中查找的 DummyService，若没有则返回固定配置的 DefaultDummyService
func GetDummyService() Service {
	dy, err := FindService(Dummy)
	if err == nil && dy != nil {
		return dy
	}
	return DefaultDummyService()
}

// DefaultDummyService 返回固定配置的（不是从 Registry 中查找的） DefaultDummyService
func DefaultDummyService() Service {
	return dummyService.Load()
}

var dummyService = xsync.OnceInit[Service]{
	New: func() Service {
		return NewDummyService(Dummy)
	},
}

// SetDefaultDummyService 替换默认的 DummyService，只影响 DefaultDummyService()
// 若 Registry 里有名字叫做 dummy 的 service，应该使用 Registry.Upsert 替换
func SetDefaultDummyService(srv Service) {
	dummyService.Store(srv)
}

func NewDummyService(name string) Service {
	cfg := &Config{
		Name:           name,
		ConnectRetry:   1,
		ConnectTimeout: xtype.Duration(10 * time.Second),
		WriteTimeout:   xtype.Duration(10 * time.Second),
		ReadTimeout:    xtype.Duration(60 * time.Second),
		Retry:          2,
		WorkerCycle:    xtype.Duration(24 * time.Hour),
		DownStream: DownStreamPart{
			Address: []string{xnet.DummyAddress},
		},
	}
	ser, err := cfg.Parser("bj")
	anygo.Must(err)
	err = ser.Start(context.Background())
	anygo.Must(err)
	return ser
}
