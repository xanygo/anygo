//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-30

package xpool_test

import (
	"context"
	"fmt"
	"io"
	"sync/atomic"
	"testing"

	"github.com/xanygo/anygo/ds/xpool"
	"github.com/xanygo/anygo/xt"
)

func TestSimple(t *testing.T) {
	ct1 := xpool.FactoryFunc[*testCloser](func(ctx context.Context) (*testCloser, error) {
		return &testCloser{
			id: tid.Add(1),
		}, nil
	})
	p1 := xpool.New[*testCloser](nil, ct1)
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("1_loop_%d", i), func(t *testing.T) {
			v1, err1 := p1.Get(context.Background())
			xt.NoError(t, err1)
			xt.NotEmpty(t, v1)
			xt.Equal(t, 1, v1.Object().id)
			v1.Release(nil)
		})
	}

	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("2_loop_%d", i), func(t *testing.T) {
			v1, err1 := p1.Get(context.Background())
			t.Log("eid=", v1.ID())
			xt.NoError(t, err1)
			xt.NotEmpty(t, v1)
			xt.Equal(t, int64(i+1), v1.Object().id)
			v1.Release(xpool.ErrBadEntry) // 放回去的时候标记错误
		})
	}
}

var _ io.Closer = (*testCloser)(nil)

var tid atomic.Int64

type testCloser struct {
	id int64
}

func (t testCloser) Close() error {
	return nil
}
