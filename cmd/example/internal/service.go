//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-22

package internal

import (
	"context"
	"fmt"
	"os"

	"github.com/xanygo/anygo"
	"github.com/xanygo/anygo/xlog"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xrpc"
	"github.com/xanygo/anygo/xnet/xservice"
)

func ServiceInit() {
	l := &xrpc.Logger{
		Logger: xlog.NewSimple(os.Stderr),
	}
	xrpc.RegisterTCPIT(l.Interceptor())

	ps := []string{"../service/*.json"}
	err1 := xservice.LoadDir(context.Background(), ps...)
	anygo.Must(err1)
	fmt.Println(" xservice.LoadDir err=", err1)
	xnet.WithInterceptor(xnet.PrintDialLogIT, xnet.PrintResolverLogIT)
}
