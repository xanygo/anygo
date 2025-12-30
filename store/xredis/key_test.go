//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-14

package xredis

import (
	"context"
	"testing"
	"time"

	"github.com/xanygo/anygo/internal/redistest"
	"github.com/xanygo/anygo/xt"
)

func TestClientKey(t *testing.T) {
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

	t.Run("TTL", func(t *testing.T) {
		val, err := client.TTL(ctx, "k1")
		xt.ErrorIs(t, err, ErrNil)
		xt.Equal(t, 0, val)

		testSetKeyString(t, client, "k1")
		val, err = client.TTL(ctx, "k1")
		xt.NoError(t, err)
		xt.Equal(t, time.Duration(-1), val)

		ok, err := client.Expire(ctx, "k1", 2*time.Second)
		xt.NoError(t, err)
		xt.True(t, ok)

		val, err = client.TTL(ctx, "k1")
		xt.NoError(t, err)
		xt.LessOrEqual(t, val, 2*time.Second)
		xt.Greater(t, val, 1*time.Second)
	})

	t.Run("PTTL", func(t *testing.T) {
		val, err := client.PTTL(ctx, "k2")
		xt.ErrorIs(t, err, ErrNil)
		xt.Equal(t, 0, val)

		testSetKeyString(t, client, "k2")
		val, err = client.PTTL(ctx, "k2")
		xt.NoError(t, err)
		xt.Equal(t, time.Duration(-1), val)

		ok, err := client.Expire(ctx, "k2", 2*time.Second)
		xt.NoError(t, err)
		xt.True(t, ok)

		val, err = client.PTTL(ctx, "k2")
		xt.NoError(t, err)
		xt.LessOrEqual(t, val, 2*time.Second)
		xt.Greater(t, val, 1*time.Second)
	})

	t.Run("Del", func(t *testing.T) {
		num, err := client.Del(ctx, "d1", "d2")
		xt.NoError(t, err)
		xt.Equal(t, 0, num)

		testSetKeyString(t, client, "d1")
		num, err = client.Del(ctx, "d1", "d2")
		xt.NoError(t, err)
		xt.Equal(t, 1, num)
	})

	t.Run("EXISTS", func(t *testing.T) {
		num, err := client.EXISTS(ctx, "e1", "e2")
		xt.NoError(t, err)
		xt.Equal(t, 0, num)

		testSetKeyString(t, client, "e1")
		num, err = client.EXISTS(ctx, "e1", "e2", "e1")
		xt.NoError(t, err)
		xt.Equal(t, 2, num)
	})

	t.Run("Touch", func(t *testing.T) {
		testDelKeys(t, client, "e1", "e2")
		num, err := client.Touch(ctx, "e1", "e2")
		xt.NoError(t, err)
		xt.Equal(t, 0, num)

		testSetKeyString(t, client, "e1")
		num, err = client.Touch(ctx, "e1", "e2")
		xt.NoError(t, err)
		xt.Equal(t, 1, num)

		vs, err := client.Keys(ctx, "e*")
		xt.NoError(t, err)
		xt.NotEmpty(t, vs)
		xt.SliceContains(t, vs, "e1")
	})

	t.Run("Move", func(t *testing.T) {
		testSetKeyString(t, client, "k1")
		ok, err := client.Move(ctx, "k1", 1)
		xt.NoError(t, err)
		xt.True(t, ok)

		ok, err = client.Move(ctx, "k1", 1)
		xt.NoError(t, err)
		xt.False(t, ok)
	})

	t.Run("Type", func(t *testing.T) {
		val, err := client.Type(ctx, "k-not-found")
		xt.ErrorIs(t, err, ErrNil)
		xt.Equal(t, "none", val)

		testSetKeyString(t, client, "k1")
		val, err = client.Type(ctx, "k1")
		xt.NoError(t, err)
		xt.Equal(t, "string", val)
	})

	t.Run("Expire", func(t *testing.T) {
		testDelKeys(t, client, "k1")
		ok, err := client.Expire(ctx, "k1", 2*time.Second)
		xt.NoError(t, err)
		xt.False(t, ok)

		testSetKeyString(t, client, "k1")
		ok, err = client.Expire(ctx, "k1", 2*time.Second)
		xt.NoError(t, err)
		xt.True(t, ok)
	})

	t.Run("ExpireAt", func(t *testing.T) {
		testDelKeys(t, client, "k1")
		ok, err := client.ExpireAt(ctx, "k1", time.Now().Add(time.Hour))
		xt.NoError(t, err)
		xt.False(t, ok)

		testSetKeyString(t, client, "k1")
		ok, err = client.ExpireAt(ctx, "k1", time.Now().Add(time.Hour))
		xt.NoError(t, err)
		xt.True(t, ok)
	})

	t.Run("ExpireTime", func(t *testing.T) {
		testDelKeys(t, client, "k1")
		tm, err := client.ExpireTime(ctx, "k1")
		xt.ErrorIs(t, err, ErrNil)
		xt.True(t, tm.IsZero())

		testSetKeyString(t, client, "k1")
		tm, err = client.ExpireTime(ctx, "k1")
		xt.NoError(t, err)
		xt.True(t, tm.IsZero())

		ok, err := client.ExpireAt(ctx, "k1", time.Now().Add(time.Hour))
		xt.NoError(t, err)
		xt.True(t, ok)

		tm, err = client.ExpireTime(ctx, "k1")
		xt.NoError(t, err)
		xt.False(t, tm.IsZero())
	})

	t.Run("PExpireTime", func(t *testing.T) {
		testDelKeys(t, client, "k1")
		tm, err := client.PExpireTime(ctx, "k1")
		xt.ErrorIs(t, err, ErrNil)
		xt.True(t, tm.IsZero())

		testSetKeyString(t, client, "k1")
		tm, err = client.PExpireTime(ctx, "k1")
		xt.NoError(t, err)
		xt.True(t, tm.IsZero())

		ok, err := client.ExpireAt(ctx, "k1", time.Now().Add(time.Hour))
		xt.NoError(t, err)
		xt.True(t, ok)

		tm, err = client.PExpireTime(ctx, "k1")
		xt.NoError(t, err)
		xt.False(t, tm.IsZero())
	})

	t.Run("RandomKey", func(t *testing.T) {
		testSetKeyString(t, client, "k1")
		k, err := client.RandomKey(ctx)
		xt.NoError(t, err)
		xt.NotEmpty(t, k)
	})

	t.Run("Rename", func(t *testing.T) {
		testSetKeyString(t, client, "k1")
		err := client.Rename(ctx, "k1", "k2")
		xt.NoError(t, err)

		err = client.Rename(ctx, "k1", "k2")
		xt.Error(t, err)
	})
	t.Run("RenameNX", func(t *testing.T) {
		testSetKeyString(t, client, "k1")
		testDelKeys(t, client, "k2", "k3")

		ok, err := client.RenameNX(ctx, "k1", "k2")
		xt.NoError(t, err)
		xt.True(t, ok)

		// k1 不存在，k3 不存在
		ok, err = client.RenameNX(ctx, "k1", "k3")
		xt.Error(t, err)
		xt.False(t, ok)
	})

	t.Run("Scan", func(t *testing.T) {
		next, keys, err := client.Scan(ctx, 0, "", 10, "")
		xt.NoError(t, err)
		xt.GreaterOrEqual(t, next, 0)
		xt.Greater(t, len(keys), 0)

		var num int
		err = client.ScanWalk(ctx, 0, "", 10, "", func(cursor uint64, keys []string) error {
			num += len(keys)
			return nil
		})
		xt.NoError(t, err)
		xt.Equal(t, len(keys), num)
	})
}

func testDelKeys(t *testing.T, client *Client, keys ...string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	_, err := client.Del(ctx, keys...)
	xt.NoError(t, err)
}
