//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-12-26

package xredis

import (
	"context"
	"testing"
	"time"

	"github.com/xanygo/anygo/internal/redistest"
	"github.com/xanygo/anygo/xt"
)

func TestClientSys(t *testing.T) {
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

	t.Run("Time", func(t *testing.T) {
		tm, err := client.Time(ctx)
		xt.NoError(t, err)
		sub := time.Since(tm)
		xt.Less(t, sub, time.Hour)
	})

	t.Run("DBSize", func(t *testing.T) {
		xt.NoError(t, client.Set(ctx, "DBSize-1", "a"))
		num, err := client.DBSize(ctx)
		xt.NoError(t, err)
		xt.NotEmpty(t, num)
	})

	t.Run("LastSave", func(t *testing.T) {
		xt.NoError(t, client.Set(ctx, "LastSave-1", "a"))

		tm, err := client.LastSave(ctx)
		xt.NoError(t, err)
		sub := time.Since(tm)
		xt.Less(t, sub, time.Hour)
	})

	t.Run("ModuleList", func(t *testing.T) {
		got, err := client.ModuleList(ctx)
		xt.NoError(t, err)
		xt.NotEmpty(t, got)
		xt.NotEmpty(t, got[0])
		xt.NotEmpty(t, got[0].Name)
		xt.NotEmpty(t, got[0].Version)
	})

	t.Run("MemoryUsage", func(t *testing.T) {
		got, err := client.MemoryUsage(ctx, "MemoryUsage-1")
		xt.ErrorIs(t, err, ErrNil)
		xt.Empty(t, got)

		xt.NoError(t, client.Set(ctx, "MemoryUsage-1", "a"))
		got, err = client.MemoryUsage(ctx, "MemoryUsage-1")
		xt.NoError(t, err)
		xt.Equal(t, 40, got)
	})

	t.Run("MemoryStats", func(t *testing.T) {
		got, err := client.MemoryStats(ctx)
		xt.NoError(t, err)
		xt.NotEmpty(t, got)
	})

	t.Run("MemoryPurge", func(t *testing.T) {
		xt.NoError(t, client.MemoryPurge(ctx))
	})

	t.Run("Info", func(t *testing.T) {
		got, err := client.Info(ctx)
		xt.NoError(t, err)
		xt.NotEmpty(t, got)
	})

	t.Run("ConfigGet", func(t *testing.T) {
		got, err := client.ConfigGet(ctx)
		xt.NoError(t, err)
		xt.NotEmpty(t, got)
	})

	t.Run("ConfigSet", func(t *testing.T) {
		got, err := client.ConfigGet(ctx, "appendonly")
		xt.NoError(t, err)
		xt.NotEmpty(t, got)

		err = client.ConfigSet(ctx, "appendonly", got["appendonly"])
		xt.NoError(t, err)
	})

	t.Run("CommandList", func(t *testing.T) {
		got, err := client.CommandList(ctx, nil)
		xt.NoError(t, err)
		xt.NotEmpty(t, got)
	})
}
