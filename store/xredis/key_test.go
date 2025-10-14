//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-14

package xredis

import (
	"context"
	"testing"
	"time"

	"github.com/fsgo/fst"

	"github.com/xanygo/anygo/internal/redistest"
)

func TestClientKey(t *testing.T) {
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

	t.Run("TTL", func(t *testing.T) {
		val, err := client.TTL(ctx, "k1")
		fst.ErrorIs(t, err, ErrNil)
		fst.Equal(t, 0, val)

		testSetKeyString(t, client, "k1")
		val, err = client.TTL(ctx, "k1")
		fst.NoError(t, err)
		fst.Equal(t, time.Duration(-1), val)

		ok, err := client.Expire(ctx, "k1", 2*time.Second)
		fst.NoError(t, err)
		fst.True(t, ok)

		val, err = client.TTL(ctx, "k1")
		fst.NoError(t, err)
		fst.LessOrEqual(t, val, 2*time.Second)
		fst.Greater(t, val, 1*time.Second)
	})

	t.Run("PTTL", func(t *testing.T) {
		val, err := client.PTTL(ctx, "k2")
		fst.ErrorIs(t, err, ErrNil)
		fst.Equal(t, 0, val)

		testSetKeyString(t, client, "k2")
		val, err = client.PTTL(ctx, "k2")
		fst.NoError(t, err)
		fst.Equal(t, time.Duration(-1), val)

		ok, err := client.Expire(ctx, "k2", 2*time.Second)
		fst.NoError(t, err)
		fst.True(t, ok)

		val, err = client.PTTL(ctx, "k2")
		fst.NoError(t, err)
		fst.LessOrEqual(t, val, 2*time.Second)
		fst.Greater(t, val, 1*time.Second)
	})

	t.Run("Del", func(t *testing.T) {
		num, err := client.Del(ctx, "d1", "d2")
		fst.NoError(t, err)
		fst.Equal(t, 0, num)

		testSetKeyString(t, client, "d1")
		num, err = client.Del(ctx, "d1", "d2")
		fst.NoError(t, err)
		fst.Equal(t, 1, num)
	})

	t.Run("EXISTS", func(t *testing.T) {
		num, err := client.EXISTS(ctx, "e1", "e2")
		fst.NoError(t, err)
		fst.Equal(t, 0, num)

		testSetKeyString(t, client, "e1")
		num, err = client.EXISTS(ctx, "e1", "e2", "e1")
		fst.NoError(t, err)
		fst.Equal(t, 2, num)
	})

	t.Run("Touch", func(t *testing.T) {
		testDelKeys(t, client, "e1", "e2")
		num, err := client.Touch(ctx, "e1", "e2")
		fst.NoError(t, err)
		fst.Equal(t, 0, num)

		testSetKeyString(t, client, "e1")
		num, err = client.Touch(ctx, "e1", "e2")
		fst.NoError(t, err)
		fst.Equal(t, 1, num)

		vs, err := client.Keys(ctx, "e*")
		fst.NoError(t, err)
		fst.NotEmpty(t, vs)
		fst.SliceContains(t, vs, "e1")
	})

	t.Run("Move", func(t *testing.T) {
		testSetKeyString(t, client, "k1")
		ok, err := client.Move(ctx, "k1", 1)
		fst.NoError(t, err)
		fst.True(t, ok)

		ok, err = client.Move(ctx, "k1", 1)
		fst.NoError(t, err)
		fst.False(t, ok)
	})

	t.Run("Type", func(t *testing.T) {
		val, err := client.Type(ctx, "k-not-found")
		fst.ErrorIs(t, err, ErrNil)
		fst.Equal(t, "none", val)

		testSetKeyString(t, client, "k1")
		val, err = client.Type(ctx, "k1")
		fst.NoError(t, err)
		fst.Equal(t, "string", val)
	})

	t.Run("Expire", func(t *testing.T) {
		testDelKeys(t, client, "k1")
		ok, err := client.Expire(ctx, "k1", 2*time.Second)
		fst.NoError(t, err)
		fst.False(t, ok)

		testSetKeyString(t, client, "k1")
		ok, err = client.Expire(ctx, "k1", 2*time.Second)
		fst.NoError(t, err)
		fst.True(t, ok)
	})

	t.Run("ExpireAt", func(t *testing.T) {
		testDelKeys(t, client, "k1")
		ok, err := client.ExpireAt(ctx, "k1", time.Now().Add(time.Hour))
		fst.NoError(t, err)
		fst.False(t, ok)

		testSetKeyString(t, client, "k1")
		ok, err = client.ExpireAt(ctx, "k1", time.Now().Add(time.Hour))
		fst.NoError(t, err)
		fst.True(t, ok)
	})

	t.Run("ExpireTime", func(t *testing.T) {
		testDelKeys(t, client, "k1")
		tm, err := client.ExpireTime(ctx, "k1")
		fst.ErrorIs(t, err, ErrNil)
		fst.True(t, tm.IsZero())

		testSetKeyString(t, client, "k1")
		tm, err = client.ExpireTime(ctx, "k1")
		fst.NoError(t, err)
		fst.True(t, tm.IsZero())

		ok, err := client.ExpireAt(ctx, "k1", time.Now().Add(time.Hour))
		fst.NoError(t, err)
		fst.True(t, ok)

		tm, err = client.ExpireTime(ctx, "k1")
		fst.NoError(t, err)
		fst.False(t, tm.IsZero())
	})

	t.Run("PExpireTime", func(t *testing.T) {
		testDelKeys(t, client, "k1")
		tm, err := client.PExpireTime(ctx, "k1")
		fst.ErrorIs(t, err, ErrNil)
		fst.True(t, tm.IsZero())

		testSetKeyString(t, client, "k1")
		tm, err = client.PExpireTime(ctx, "k1")
		fst.NoError(t, err)
		fst.True(t, tm.IsZero())

		ok, err := client.ExpireAt(ctx, "k1", time.Now().Add(time.Hour))
		fst.NoError(t, err)
		fst.True(t, ok)

		tm, err = client.PExpireTime(ctx, "k1")
		fst.NoError(t, err)
		fst.False(t, tm.IsZero())
	})

	t.Run("RandomKey", func(t *testing.T) {
		testSetKeyString(t, client, "k1")
		k, err := client.RandomKey(ctx)
		fst.NoError(t, err)
		fst.NotEmpty(t, k)
	})

	t.Run("Rename", func(t *testing.T) {
		testSetKeyString(t, client, "k1")
		err := client.Rename(ctx, "k1", "k2")
		fst.NoError(t, err)

		err = client.Rename(ctx, "k1", "k2")
		fst.Error(t, err)
	})
	t.Run("RenameNX", func(t *testing.T) {
		testSetKeyString(t, client, "k1")
		testDelKeys(t, client, "k2", "k3")

		ok, err := client.RenameNX(ctx, "k1", "k2")
		fst.NoError(t, err)
		fst.True(t, ok)

		// k1 不存在，k3 不存在
		ok, err = client.RenameNX(ctx, "k1", "k3")
		fst.Error(t, err)
		fst.False(t, ok)
	})
}

func testDelKeys(t *testing.T, client *Client, keys ...string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	_, err := client.Del(ctx, keys...)
	fst.NoError(t, err)
}
