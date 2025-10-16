//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-16

package xredis

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/fsgo/fst"

	"github.com/xanygo/anygo/internal/redistest"
)

func TestClientZSet(t *testing.T) {
	ts, errTs := redistest.NewServer()
	if errTs != nil {
		t.Logf("create redis fail: %v", errTs)
		return
	}
	defer ts.Stop()
	t.Logf("uri= %q", ts.URI())

	_, client, errClient := NewClientByURI("demo", ts.URI())
	fst.NoError(t, errClient)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	t.Run("ZAdd", func(t *testing.T) {
		ok, err := client.ZAdd(ctx, "z1", 1, "f1")
		fst.NoError(t, err)
		fst.True(t, ok)

		ok, err = client.ZAdd(ctx, "z1", 1, "f1")
		fst.NoError(t, err)
		fst.False(t, ok)
		ok, err = client.ZAdd(ctx, "z1", math.Inf(-1), "f1")
		fst.NoError(t, err)
		fst.False(t, ok)
	})

	t.Run("ZAddOpt", func(t *testing.T) {
		ok, err := client.ZAdd(ctx, "z2", 1, "f1")
		fst.NoError(t, err)
		fst.True(t, ok)

		num, err := client.ZAddOpt(ctx, "z2", []string{"NX"}, 1, "f1")
		fst.NoError(t, err)
		fst.Equal(t, 0, num)
	})

	data := map[string]float64{
		"f1": 1,
		"f2": 2,
	}
	t.Run("ZAddMap", func(t *testing.T) {
		num, err := client.ZAddMap(ctx, "z3", data)
		fst.NoError(t, err)
		fst.Equal(t, 2, num)
	})

	t.Run("ZAddMapOpt", func(t *testing.T) {
		num, err := client.ZAddMapOpt(ctx, "z4", []string{"NX"}, data)
		fst.NoError(t, err)
		fst.Equal(t, 2, num)

		num, err = client.ZAddMapOpt(ctx, "z4", []string{"NX"}, data)
		fst.NoError(t, err)
		fst.Equal(t, 0, num)

		num, err = client.ZCard(ctx, "z4")
		fst.NoError(t, err)
		fst.Equal(t, 2, num)
	})

	t.Run("ZCount", func(t *testing.T) {
		num, err := client.ZCount(ctx, "z4", 0, math.Inf(1))
		fst.NoError(t, err)
		fst.Greater(t, num, 0)
	})
}
