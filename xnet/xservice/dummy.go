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

func DummyService() Service {
	return dummyService.Load()
}

var dummyService = xsync.OnceInit[Service]{
	New: func() Service {
		return NewDummyService(Dummy)
	},
}

func NewDummyService(name string) Service {
	cfg := &Config{
		Name:           name,
		ConnectRetry:   1,
		ConnectTimeout: xtype.Duration(5000 * time.Millisecond),
		WriteTimeout:   xtype.Duration(10000 * time.Millisecond),
		ReadTimeout:    xtype.Duration(45000 * time.Millisecond),
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
