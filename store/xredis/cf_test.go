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

func TestClientCF(t *testing.T) {
	ts, errTs := redistest.NewServer()
	if errTs != nil {
		t.Skipf("create redis-server skipped: %v", errTs)
		return
	}
	defer ts.Stop()
	t.Logf("uri= %q", ts.URI())
	_, client, errClient := NewClientByURI("demo", ts.URI())
	xt.NoError(t, errClient)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	t.Run("CFAdd", func(t *testing.T) {
		got, err := client.CFAdd(ctx, "CFAdd-1", "world")
		xt.NoError(t, err)
		xt.True(t, got)
	})

	t.Run("CFAddNX", func(t *testing.T) {
		got, err := client.CFAddNX(ctx, "CFAddNX-1", "world")
		xt.NoError(t, err)
		xt.True(t, got)

		got, err = client.CFAddNX(ctx, "CFAddNX-1", "world")
		xt.NoError(t, err)
		xt.False(t, got)

		count, err := client.CFCount(ctx, "CFAddNX-1", "world")
		xt.NoError(t, err)
		xt.Equal(t, 1, count)

		count, err = client.CFCount(ctx, "CFAddNX-1", "not-found")
		xt.NoError(t, err)
		xt.Equal(t, 0, count)

		count, err = client.CFCount(ctx, "CFAddNX-1-not-found", "not-found")
		xt.NoError(t, err)
		xt.Equal(t, 0, count)
	})

	t.Run("CFDel", func(t *testing.T) {
		got, err := client.CFDel(ctx, "CFDel-1", "world")
		xt.Error(t, err) // 返回 error("ERR not found")
		xt.False(t, got)

		_, err = client.CFAdd(ctx, "CFDel-1", "world")
		xt.NoError(t, err)

		got, err = client.CFDel(ctx, "CFDel-1", "world")
		xt.NoError(t, err)
		xt.True(t, got)

		// 再次删除
		got, err = client.CFDel(ctx, "CFDel-1", "world")
		xt.NoError(t, err)
		xt.False(t, got)
	})

	t.Run("CFExists", func(t *testing.T) {
		got, err := client.CFExists(ctx, "CFExists-1", "world")
		xt.NoError(t, err)
		xt.False(t, got)

		_, err = client.CFAdd(ctx, "CFExists-1", "world")
		xt.NoError(t, err)

		got, err = client.CFExists(ctx, "CFExists-1", "world")
		xt.NoError(t, err) // 返回 error("ERR not found")
		xt.True(t, got)
	})

	t.Run("CFInFo", func(t *testing.T) {
		got, err := client.CFInFo(ctx, "CFInFo-1")
		xt.Error(t, err)
		xt.Empty(t, got)

		_, err = client.CFAdd(ctx, "CFInFo-1", "world")
		xt.NoError(t, err)

		got, err = client.CFInFo(ctx, "CFInFo-1")
		xt.NoError(t, err)
		xt.NotEmpty(t, got)
	})

	t.Run("CFInsert", func(t *testing.T) {
		got, err := client.CFInsert(ctx, "CFInsert-1", "k1", "k2")
		xt.NoError(t, err)
		xt.Equal(t, []bool{true, true}, got)

		// 重复添加一次
		got, err = client.CFInsertWithOption(ctx, "CFInsert-1", nil, "k1", "k2")
		xt.NoError(t, err)
		xt.Equal(t, []bool{true, true}, got)
	})

	t.Run("CFInsertNX", func(t *testing.T) {
		got, err := client.CFInsertNX(ctx, "CFInsertNX-1", "k1", "k2")
		xt.NoError(t, err)
		xt.Equal(t, []bool{true, true}, got)

		// 重复添加一次: 全部失败
		got, err = client.CFInsertNXWithOption(ctx, "CFInsertNX-1", nil, "k1", "k2")
		xt.NoError(t, err)
		xt.Equal(t, []bool{false, false}, got)

		// 重复添加一次: 部分成功
		got, err = client.CFInsertNXWithOption(ctx, "CFInsertNX-1", nil, "k1", "k2", "k3")
		xt.NoError(t, err)
		xt.Equal(t, []bool{false, false, true}, got)
	})

	t.Run("CFMExists", func(t *testing.T) {
		got, err := client.CFMExists(ctx, "CFMExists-1", "k1", "k2")
		xt.NoError(t, err)
		xt.Equal(t, []bool{false, false}, got)

		_, err = client.CFAdd(ctx, "CFMExists-1", "k1")
		xt.NoError(t, err)

		got, err = client.CFMExists(ctx, "CFMExists-1", "k1", "k2")
		xt.NoError(t, err)
		xt.Equal(t, []bool{true, false}, got)
	})

	t.Run("CFReserve", func(t *testing.T) {
		err := client.CFReserve(ctx, "CFReserve-1", 100, nil)
		xt.NoError(t, err)

		err = client.CFReserve(ctx, "CFReserve-1", 100, nil)
		xt.Error(t, err) // 重复创建，报错

		err = client.CFReserve(ctx, "CFReserve-2", 100, &CFReserveOption{Expansion: 2, BucketSize: 50, MaxIterations: 30})
		xt.NoError(t, err)
	})

	t.Run("CFScanDump", func(t *testing.T) {
		_, err := client.CFAdd(ctx, "CFScanDump-1", "k1")
		xt.NoError(t, err)
		next, data, err := client.CFScanDump(ctx, "CFScanDump-1", 0)
		xt.NoError(t, err)
		xt.NotEmpty(t, data)
		xt.GreaterOrEqual(t, next, 0)

		err = client.CFLoadChunk(ctx, "CFLoadChunk-1", next, data)
		xt.NoError(t, err)
	})
}
