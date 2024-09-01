//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-31

package zcache

import (
	"context"
	"fmt"
	"testing"

	"github.com/fsgo/fst"
)

func TestMap(t *testing.T) {
	cache := &Map[int, int]{
		New: func(ctx context.Context, key int) (int, error) {
			if key < 100 {
				return key + 5, nil
			}
			return 0, fmt.Errorf("invalid key %d", key)
		},
		Caption: 10,
	}
	got1, err1 := cache.Get(context.Background(), 1)
	fst.Equal(t, 6, got1)
	fst.NoError(t, err1)
}
