//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-12-24

package xredis

import (
	"context"
	"testing"
	"time"

	"github.com/xanygo/anygo/internal/redistest"
	"github.com/xanygo/anygo/xt"
)

func TestClientBF(t *testing.T) {
	ts, errTs := redistest.NewServer()
	if errTs != nil {
		t.Logf("create redis fail: %v", errTs)
		return
	}
	defer ts.Stop()
	t.Logf("uri= %q", ts.URI())
	_, client, errClient := NewClientByURI("demo", ts.URI())
	xt.NoError(t, errClient)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	bfAdd := func(t *testing.T, key string, value string) {
		got, err := client.BFAdd(ctx, key, value)
		xt.NoError(t, err)
		xt.True(t, got)
	}

	t.Run("BFAdd", func(t *testing.T) {
		got, err := client.BFAdd(ctx, "bf1", "v1")
		xt.NoError(t, err)
		xt.True(t, got)

		got, err = client.BFAdd(ctx, "bf1", "v1")
		xt.NoError(t, err)
		xt.False(t, got)

		xt.NoError(t, client.Set(ctx, "bf1", "v"))

		got, err = client.BFAdd(ctx, "bf1", "v2")
		xt.Error(t, err)
		xt.False(t, got)
	})

	t.Run("BFMAdd", func(t *testing.T) {
		got, err := client.BFMAdd(ctx, "bf-madd", "v1")
		xt.NoError(t, err)
		xt.Equal(t, []bool{true}, got)

		got, err = client.BFMAdd(ctx, "bf-madd", "v1")
		xt.NoError(t, err)
		xt.Equal(t, []bool{false}, got)

		got, err = client.BFMAdd(ctx, "bf-madd", "v1", "v2")
		xt.NoError(t, err)
		xt.Equal(t, []bool{false, true}, got)
	})

	t.Run("BFInfo", func(t *testing.T) {
		info, err := client.BFInfo(ctx, "bf2")
		xt.Error(t, err)
		xt.Empty(t, info)

		bfAdd(t, "bf2", "v1")

		info, err = client.BFInfo(ctx, "bf2")
		xt.NoError(t, err)
		xt.NotEmpty(t, info)
		t.Logf("BFInfo=%#v", info)
		xt.Equal(t, 1, info.Items)
		xt.Equal(t, 1, info.Filters)
		xt.Equal(t, 2, info.Expansion)
		xt.Greater(t, info.Capacity, 10)
		xt.Greater(t, info.Size, 10)
	})

	t.Run("BFCard", func(t *testing.T) {
		got, err := client.BFCard(ctx, "bf-card") // not exists
		xt.NoError(t, err)
		xt.Equal(t, 0, got)

		bfAdd(t, "bf-card", "v1")

		got, err = client.BFCard(ctx, "bf-card") // not exists
		xt.NoError(t, err)
		xt.Equal(t, 1, got)
	})

	t.Run("BFExists", func(t *testing.T) {
		got, err := client.BFExists(ctx, "bf-exists", "v1") // not exists
		xt.NoError(t, err)
		xt.False(t, got)

		got1, err1 := client.BFMExists(ctx, "bf-exists", "v1")
		xt.NoError(t, err1)
		xt.Equal(t, []bool{false}, got1)

		bfAdd(t, "bf-exists", "v1")

		got, err = client.BFExists(ctx, "bf-exists", "v1") //  exists
		xt.NoError(t, err)
		xt.True(t, got)

		got1, err1 = client.BFMExists(ctx, "bf-exists", "v1")
		xt.NoError(t, err1)
		xt.Equal(t, []bool{true}, got1)
	})

	t.Run("BFInsert", func(t *testing.T) {
		got, err := client.BFInsert(ctx, "bf-insert1", []string{"v1"}, nil)
		xt.NoError(t, err)
		xt.Equal(t, []bool{true}, got)

		got, err = client.BFInsert(ctx, "bf-insert1", []string{"v1", "v2"}, nil)
		xt.NoError(t, err)
		xt.Equal(t, []bool{false, true}, got)
	})

	t.Run("BFReserve", func(t *testing.T) {
		err := client.BFReserve(ctx, "bf-BFReserve", 0.001, 100, nil)
		xt.NoError(t, err)

		err = client.BFReserve(ctx, "bf-BFReserve", 0.001, 100, nil)
		xt.Error(t, err)

		err = client.BFReserve(ctx, "bf-BFReserve-1", 0.001, 1000, &BFReserveOption{
			NonScaling: true,
		})
		xt.NoError(t, err)

		err = client.BFReserve(ctx, "bf-BFReserve-2", 0.001, 1000, &BFReserveOption{
			Expansion: 1,
		})
		xt.NoError(t, err)

		// 同时设置会报错
		err = client.BFReserve(ctx, "bf-BFReserve-3", 0.001, 1000, &BFReserveOption{
			Expansion:  1,
			NonScaling: true,
		})
		xt.Error(t, err)
	})
}
