//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-30

package xpool_test

import (
	"context"
	"io"
	"testing"

	"github.com/fsgo/fst"

	"github.com/xanygo/anygo/ds/xpool"
)

func TestSimple(t *testing.T) {
	ct1 := xpool.FactoryFunc[io.Closer](func(ctx context.Context) (io.Closer, error) {
		return &testCloser{}, nil
	})
	p1 := xpool.New[io.Closer](nil, ct1)
	v1, err1 := p1.Get(context.Background())
	fst.NoError(t, err1)
	fst.NotEmpty(t, v1)
}

var _ io.Closer = (*testCloser)(nil)

type testCloser struct {
}

func (t testCloser) Close() error {
	return nil
}
