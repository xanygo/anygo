//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-02

package xctx

import (
	"context"
	"testing"

	"github.com/fsgo/fst"
)

type tk uint8

const tk0 tk = iota

func TestWithValues(t *testing.T) {
	ctx1 := WithValues(context.Background(), tk0, 1, 2)
	ctx2 := WithValues(ctx1, tk0, 3)

	vs1 := Values[tk, int](ctx2, tk0, true)
	fst.Equal(t, []int{1, 2, 3}, vs1)

	vs2 := Values[tk, int](ctx2, tk0, false)
	fst.Equal(t, []int{3}, vs2)

	vs3 := Values[tk, int](context.Background(), tk0, false)
	fst.Empty(t, vs3)
}
