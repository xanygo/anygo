//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-22

package xtp_test

import (
	"context"
	"fmt"
	"time"

	"github.com/xanygo/anygo/xtp"
)

func ExampleHedging_Run() {
	hg := &xtp.Hedging[int]{
		Main: func(ctx context.Context) (int, error) {
			// 模拟长耗时 1 秒
			select {
			case <-ctx.Done():
			case <-time.After(time.Second):
			}
			return 1, nil
		},
	}
	hg.Add(10*time.Microsecond, func(ctx context.Context) (int, error) {
		return 2, nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	got, err := hg.Run(ctx)
	fmt.Println("got=", got, ", err=", err)

	// Output:
	// got= 2 , err= <nil>
}
