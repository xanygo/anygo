//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-13

package xredis

import (
	"context"
	"testing"
	"time"

	"github.com/fsgo/fst"

	"github.com/xanygo/anygo/internal/redistest"
)

func TestClientString(t *testing.T) {
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

	t.Run("get", func(t *testing.T) {
		value, err := client.Get(ctx, "k1")
		fst.ErrorIs(t, err, ErrNil)
		fst.Empty(t, value)
	})

	t.Run("set", func(t *testing.T) {
		err := client.Set(ctx, "k1", "v1", 0)
		fst.NoError(t, err)
		value, err := client.Get(ctx, "k1")
		fst.NoError(t, err)
		fst.Equal(t, "v1", value)

		num, err := client.Del(ctx, "k1", "k2")
		fst.NoError(t, err)
		fst.Equal(t, 1, num)

		value, err = client.Get(ctx, "k1")
		fst.ErrorIs(t, err, ErrNil)
		fst.Empty(t, value)

		err = client.Set(ctx, "k1", "v1", time.Millisecond)
		fst.NoError(t, err)
		time.Sleep(2 * time.Millisecond)

		value, err = client.Get(ctx, "k1")
		fst.ErrorIs(t, err, ErrNil)
		fst.Empty(t, value)
	})

	t.Run("setnx", func(t *testing.T) {
		ok, err := client.SetNX(ctx, "k2", "v2", 0)
		fst.NoError(t, err)
		fst.True(t, ok)

		ok, err = client.SetNX(ctx, "k2", "v2-1", 0)
		fst.NoError(t, err)
		fst.False(t, ok)
	})

	t.Run("setxx", func(t *testing.T) {
		ok, err := client.SetXX(ctx, "k3", "v3", 0)
		fst.NoError(t, err)
		fst.False(t, ok)

		_, err = client.Get(ctx, "k3")
		fst.ErrorIs(t, err, ErrNil)

		err = client.Set(ctx, "k3", "v3", 0)
		fst.NoError(t, err)

		ok, err = client.SetXX(ctx, "k3", "v3", 0)
		fst.NoError(t, err)
		fst.True(t, ok)
	})

	t.Run("incr_decr", func(t *testing.T) {
		num, err := client.Incr(ctx, "k3") // invalid type
		fst.Error(t, err)
		fst.Equal(t, 0, num)

		num, err = client.Incr(ctx, "k4")
		fst.NoError(t, err)
		fst.Equal(t, 1, num)

		num, err = client.IncrBy(ctx, "k4", 3)
		fst.NoError(t, err)
		fst.Equal(t, 4, num)

		num, err = client.Decr(ctx, "k4")
		fst.NoError(t, err)
		fst.Equal(t, 3, num)

		num, err = client.DecrBy(ctx, "k4", 2)
		fst.NoError(t, err)
		fst.Equal(t, 1, num)
	})

	t.Run("getdel", func(t *testing.T) {
		val, err := client.GetDel(ctx, "k4")
		fst.NoError(t, err)
		fst.Equal(t, "1", val)

		val, err = client.GetDel(ctx, "k4")
		fst.ErrorIs(t, err, ErrNil)
		fst.Equal(t, "", val)
	})

	t.Run("getset", func(t *testing.T) {
		val, err := client.GetSet(ctx, "k5", "v5")
		fst.ErrorIs(t, err, ErrNil)
		fst.Equal(t, "", val)

		val, err = client.GetSet(ctx, "k5", "v5-1")
		fst.NoError(t, err)
		fst.Equal(t, "v5", val)

		val, err = client.GetSet(ctx, "k5", "v5-2")
		fst.NoError(t, err)
		fst.Equal(t, "v5-1", val)
	})

	t.Run("getrange", func(t *testing.T) {
		val, err := client.GetRange(ctx, "k6", 0, 4)
		fst.NoError(t, err)
		fst.Equal(t, "", val)

		fst.NoError(t, client.Set(ctx, "k6", "hello-world", 0))

		val, err = client.GetRange(ctx, "k6", 0, 4)
		fst.NoError(t, err)
		fst.Equal(t, "hello", val)

		val, err = client.GetRange(ctx, "k6", 0, 100)
		fst.NoError(t, err)
		fst.Equal(t, "hello-world", val)
	})

	t.Run("echo", func(t *testing.T) {
		val, err := client.Echo(ctx, "hello")
		fst.NoError(t, err)
		fst.Equal(t, "hello", val)
	})
}

func testSetKeyString(t *testing.T, client *Client, key string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	_, err := client.Del(ctx, key)
	fst.NoError(t, err)
	err = client.Set(ctx, key, "str", 0)
	fst.NoError(t, err)
}
