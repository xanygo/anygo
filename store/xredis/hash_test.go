//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-14

package xredis

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/internal/redistest"
	"github.com/xanygo/anygo/xt"
)

func TestClientHash(t *testing.T) {
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

	t.Run("HSet", func(t *testing.T) {
		num, err := client.HSet(ctx, "h1", "f1", "v1")
		xt.NoError(t, err)
		xt.Equal(t, 1, num)

		num, err = client.HSet(ctx, "h1", "f1", "v1")
		xt.NoError(t, err)
		xt.Equal(t, 0, num)

		testSetKeyString(t, client, "h1")

		num, err = client.HSet(ctx, "h1", "f1", "v1")
		xt.Error(t, err)
		xt.Equal(t, 0, num)
	})

	t.Run("HSetMap", func(t *testing.T) {
		num, err := client.HSetMap(ctx, "h2", map[string]string{"f1": "v1", "f2": "v2"})
		xt.NoError(t, err)
		xt.Equal(t, 2, num)
	})

	t.Run("HSetEX", func(t *testing.T) {
		data := map[string]string{"f1": "v1", "f2": "v2"}
		ok, err := client.HSetEX(ctx, "h3", "", -1, data)
		xt.NoError(t, err)
		xt.True(t, ok)

		ok, err = client.HSetEX(ctx, "h3", "FNX", -1, data)
		xt.NoError(t, err)
		xt.False(t, ok)

		testSetKeyString(t, client, "h3")

		ok, err = client.HSetEX(ctx, "h3", "", -1, data)
		xt.Error(t, err)
		xt.False(t, ok)
	})

	t.Run("HSetNX", func(t *testing.T) {
		ok, err := client.HSetNX(ctx, "h4", "f1", "v1")
		xt.NoError(t, err)
		xt.True(t, ok)

		ok, err = client.HSetNX(ctx, "h4", "f1", "v1")
		xt.NoError(t, err)
		xt.False(t, ok)

		testSetKeyString(t, client, "h4")

		ok, err = client.HSetNX(ctx, "h4", "f1", "v1")
		xt.Error(t, err)
		xt.False(t, ok)
	})

	t.Run("HStrLen", func(t *testing.T) {
		_, err := client.HSet(ctx, "h5", "f1", "v1")
		xt.NoError(t, err)

		num, err := client.HStrLen(ctx, "h5", "f1")
		xt.NoError(t, err)
		xt.Equal(t, 2, num)

		num, err = client.HDel(ctx, "h5", "f1", "f2")
		xt.NoError(t, err)
		xt.Equal(t, 1, num)
	})

	t.Run("HExists", func(t *testing.T) {
		ok, err := client.HExists(ctx, "h2", "f1")
		xt.NoError(t, err)
		xt.True(t, ok)
		testSetKeyString(t, client, "h2")

		ok, err = client.HExists(ctx, "h2", "f1")
		xt.Error(t, err)
		xt.False(t, ok)
	})

	t.Run("HGet", func(t *testing.T) {
		_, err := client.HSet(ctx, "h6", "f1", "v1")
		xt.NoError(t, err)

		val, err := client.HGet(ctx, "h6", "f1")
		xt.NoError(t, err)
		xt.Equal(t, "v1", val)

		val, err = client.HGet(ctx, "h6", "f2-not-exist")
		xt.ErrorIs(t, err, ErrNil)
		xt.Equal(t, "", val)
	})

	t.Run("HGetAll", func(t *testing.T) {
		data := map[string]string{"f1": "v1", "f2": "v2"}
		_, err := client.HSetMap(ctx, "HGetAll-1", data)
		xt.NoError(t, err)

		val, err := client.HGetAll(ctx, "HGetAll-1")
		xt.NoError(t, err)
		xt.Equal(t, data, val)

		val, err = client.HGetAll(ctx, "h6-not-exists")
		xt.NoError(t, err)
		xt.Empty(t, val)

		testSetKeyString(t, client, "HGetAll-1")

		val, err = client.HGetAll(ctx, "HGetAll-1")
		xt.Error(t, err)
		xt.Empty(t, val)
	})

	t.Run("HGetDel", func(t *testing.T) {
		data := map[string]string{"f1": "v1", "f2": "v2"}
		_, err := client.HSetMap(ctx, "h7", data)
		xt.NoError(t, err)

		val, err := client.HGetDel(ctx, "h7", "f1")
		xt.NoError(t, err)
		xt.Equal(t, map[string]string{"f1": "v1"}, val)

		val, err = client.HGetDel(ctx, "h7", "f1", "f2")
		xt.NoError(t, err)
		xt.Equal(t, map[string]string{"f2": "v2"}, val)
	})

	t.Run("HPersist", func(t *testing.T) {
		data := map[string]string{"f1": "v1", "f2": "v2"}
		_, err := client.HSetMap(ctx, "h7", data)
		xt.NoError(t, err)

		// todo
	})

	t.Run("HIncrBy", func(t *testing.T) {
		num, err := client.HIncrBy(ctx, "h8", "f1", 2)
		xt.NoError(t, err)
		xt.Equal(t, 2, num)

		testSetKeyString(t, client, "h8")
		num, err = client.HIncrBy(ctx, "h8", "f1", 2)
		xt.Error(t, err)
		xt.Equal(t, 0, num)
	})

	t.Run("HIncrFloat", func(t *testing.T) {
		num, err := client.HIncrFloat(ctx, "h9", "f1", 2)
		xt.NoError(t, err)
		xt.Equal(t, 2.0, num)

		keys, err := client.HKeys(ctx, "h9")
		xt.NoError(t, err)
		xt.Equal(t, []string{"f1"}, keys)

		num1, err := client.HLen(ctx, "h9")
		xt.NoError(t, err)
		xt.Equal(t, 1, num1)

		testSetKeyString(t, client, "h9")
		num, err = client.HIncrFloat(ctx, "h9", "f1", 2)
		xt.Error(t, err)
		xt.Equal(t, 0, num)
	})

	t.Run("HKeys", func(t *testing.T) {
		keys, err := client.HKeys(ctx, "h10-not-exists")
		xt.NoError(t, err)
		xt.Empty(t, keys)

		testSetKeyString(t, client, "h10")

		keys, err = client.HKeys(ctx, "h10")
		xt.Error(t, err)
		xt.Empty(t, keys)
	})

	t.Run("HMGet", func(t *testing.T) {
		data := map[string]string{"f1": "v1", "f2": "v2"}
		_, err := client.HSetMap(ctx, "h11", data)
		xt.NoError(t, err)

		val, err := client.HMGet(ctx, "h11", "f1", "f2", "f3")
		xt.NoError(t, err)
		xt.Equal(t, data, val)

		val, err = client.HMGet(ctx, "h11-not-exists", "f1", "f2", "f3")
		xt.NoError(t, err)
		xt.Empty(t, val)
	})

	t.Run("HTTL", func(t *testing.T) {
		got, err := client.HTTL(ctx, "HTTL-1", "f1", "f2")
		xt.NoError(t, err)
		xt.Len(t, got, 2)
		xt.Equal(t, []time.Duration{-2, -2}, got)

		data := map[string]string{"f1": "v1", "f2": "v2"}
		_, err = client.HSetMap(ctx, "HTTL-1", data)
		xt.NoError(t, err)

		got, err = client.HTTL(ctx, "HTTL-1", "f1", "f2")
		xt.NoError(t, err)
		xt.Len(t, got, 2)
		xt.Equal(t, []time.Duration{-1, -1}, got)

		nums, err := client.HExpireAt(ctx, "HTTL-1", time.Now().Add(time.Hour), "", "f1")
		xt.NoError(t, err)
		xt.Len(t, nums, 1)
		xt.Equal(t, 1, nums[0])

		got, err = client.HTTL(ctx, "HTTL-1", "f1", "f2")
		xt.NoError(t, err)
		xt.Len(t, got, 2)
		xt.Greater(t, got[0], 50*time.Minute)
		xt.Equal(t, time.Duration(-1), got[1])
	})

	t.Run("HVals", func(t *testing.T) {
		vals, err := client.HVals(ctx, "HVals-1")
		xt.NoError(t, err)
		xt.Empty(t, vals)

		data := map[string]string{"f1": "v1", "f2": "v2"}
		_, err = client.HSetMap(ctx, "HVals-1", data)
		xt.NoError(t, err)

		vals, err = client.HVals(ctx, "HVals-1")
		xt.NoError(t, err)
		slices.Sort(vals)
		xt.Equal(t, []string{"v1", "v2"}, vals)
	})

	t.Run("HScan", func(t *testing.T) {
		next, vals, err := client.HScan(ctx, "HScan-1", 0, "", 10)
		xt.NoError(t, err)
		xt.Empty(t, vals)
		xt.Equal(t, 0, next)

		data := map[string]string{"f1": "v1", "f2": "v2"}
		_, err = client.HSetMap(ctx, "HScan-1", data)
		xt.NoError(t, err)

		next, vals, err = client.HScan(ctx, "HScan-1", 0, "", 10)
		xt.NoError(t, err)
		xt.Equal(t, data, vals)
		xt.Equal(t, 0, next)

		var count int
		err = client.HScanWalk(ctx, "HScan-1", 0, "", 10, func(cursor uint64, data map[string]string) error {
			count += len(data)
			return nil
		})
		xt.NoError(t, err)
		xt.Equal(t, len(vals), count)
	})

	t.Run("HScanNoValues", func(t *testing.T) {
		next, vals, err := client.HScanNoValues(ctx, "HScanNoValues-1", 0, "", 10)
		xt.NoError(t, err)
		xt.Empty(t, vals)
		xt.Equal(t, 0, next)

		data := map[string]string{"f1": "v1", "f2": "v2"}
		_, err = client.HSetMap(ctx, "HScanNoValues-1", data)
		xt.NoError(t, err)

		next, vals, err = client.HScanNoValues(ctx, "HScanNoValues-1", 0, "", 10)
		xt.NoError(t, err)
		xt.SliceSortEqual(t, xmap.Keys(data), vals)
		xt.Equal(t, 0, next)

		var count int
		err = client.HScanNoValuesWalk(ctx, "HScan-1", 0, "", 10, func(cursor uint64, data []string) error {
			count += len(data)
			return nil
		})
		xt.NoError(t, err)
		xt.Equal(t, len(vals), count)
	})
}
