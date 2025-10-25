//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-29

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/xanygo/anygo/store/xredis"
	"github.com/xanygo/anygo/xlog"
	"github.com/xanygo/anygo/xnet/xrpc"
	"github.com/xanygo/anygo/xnet/xservice"
)

func main() {
	l := &xrpc.Logger{
		Logger: xlog.NewSimple(os.Stderr),
	}
	xrpc.RegisterTCPIT(l.Interceptor())

	xservice.LoadDir(context.Background(), "../service/*")
	rc := xredis.NewClient("rds")
	sr, err := rc.ZAdd(context.Background(), "z1", 1, "f2")
	fmt.Println("ret=", sr, "err=", err)
}
