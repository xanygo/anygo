//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-31

package xsession

import (
	"context"
	"testing"
	"time"

	"github.com/fsgo/fst"

	"github.com/xanygo/anygo/xerror"
)

func TestNewMemoryStore(t *testing.T) {
	ms := NewMemoryStore(100, time.Minute)
	got1, err1 := ms.Get(context.Background(), "123")
	fst.ErrorIs(t, err1, xerror.NotFound)
	fst.Nil(t, got1)

	got2 := ms.GetOrCreate(context.Background(), "123")
	fst.NotNil(t, got2)
	fst.Equal(t, "123", got2.ID())

	got2.Set("k1", "v1")
	fst.Equal(t, "v1", got2.Get("k1"))
	got2.Delete("k1")
	fst.Empty(t, got2.Get("k1"))
	fst.NoError(t, got2.Save(context.Background()))

	got2.Set("k2", "v2")
	fst.False(t, got2.CompareAndDelete("k2", "v1"))
	fst.True(t, got2.CompareAndDelete("k2", "v2"))
	fst.False(t, got2.Has("k2"))
}
