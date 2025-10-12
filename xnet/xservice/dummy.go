//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-12

package xservice

import (
	"context"
	"time"

	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xpp"
)

const (
	Dummy = xnet.Dummy
)

func DummyService() Service {
	return dummyService.Load()
}

var dummyService = xsync.OnceInit[Service]{
	New: func() Service {
		ser, err := dummyServiceConfig.Parser("bj")
		if err != nil {
			panic(err)
		}
		xpp.TryStartWorker(context.Background(), time.Hour, ser)
		return ser
	},
}

var dummyServiceConfig = &Config{
	Name:           Dummy,
	ConnectRetry:   1,
	ConnectTimeout: 5000,
	WriteTimeout:   5000,
	ReadTimeout:    5000,
	Retry:          2,
	DownStream: DownStreamPart{
		Address: []string{xnet.DummyAddress},
	},
}
