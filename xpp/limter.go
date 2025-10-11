//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-25

package xpp

import (
	"context"
	"sync"
)

// NewConcLimiter 创建新的 ConcLimiter 实例
//
// max: 最大并发度,应 >= 0。当 max=0 时，并发度无限制。
func NewConcLimiter(max int) *ConcLimiter {
	return &ConcLimiter{
		max: max,
		sem: make(chan struct{}, max),
	}
}

// ConcLimiter 并发度限制器
type ConcLimiter struct {
	sem chan struct{}

	// Max 最大并发度。若值 <1,则无限制
	max int
}

// Wait 同步获取令牌
//
// 返回的 func() 用于释放令牌
func (c *ConcLimiter) Wait() func() {
	release, _ := c.WaitContext(context.Background())
	return release
}

// WaitContext 获取令牌，若失败会返回 error
//
// 返回的第一个 func() 用于释放令牌
func (c *ConcLimiter) WaitContext(ctx context.Context) (func(), error) {
	if c.max < 1 {
		return empty, nil
	}

	select {
	case <-ctx.Done():
		return nil, context.Cause(ctx)
	case c.sem <- struct{}{}:
		return sync.OnceFunc(c.release), nil
	}
}

func (c *ConcLimiter) release() {
	<-c.sem
}

func empty() {}
