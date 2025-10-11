//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-22

package xpp_test

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/xanygo/anygo/xpp"
)

func ExampleHedging_Run() {
	hg := &xpp.Hedging[int]{
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

func ExampleConcLimiter_Wait() {
	limiter := xpp.NewConcLimiter(1) // 并发度为 1

	var wg sync.WaitGroup
	start := time.Now()
	for i := 0; i < 3; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			fn := limiter.Wait() // 获取令牌
			defer fn()           // 释放令牌

			time.Sleep(10 * time.Millisecond)
		}(i)
	}
	wg.Wait()
	cost := time.Since(start)
	fmt.Println(cost >= 30*time.Millisecond) // true

	// Output:
	// true
}
