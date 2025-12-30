//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-13

package xredis

import (
	"context"
	"testing"
	"time"

	"github.com/xanygo/anygo/internal/redistest"
	"github.com/xanygo/anygo/xt"
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
	xt.NoError(t, errClient)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	t.Run("get", func(t *testing.T) {
		value, err := client.Get(ctx, "k1")
		xt.ErrorIs(t, err, ErrNil)
		xt.Empty(t, value)
	})

	t.Run("set", func(t *testing.T) {
		err := client.Set(ctx, "k1", "v1")
		xt.NoError(t, err)
		value, err := client.Get(ctx, "k1")
		xt.NoError(t, err)
		xt.Equal(t, "v1", value)

		num, err := client.Del(ctx, "k1", "k2")
		xt.NoError(t, err)
		xt.Equal(t, 1, num)

		value, err = client.Get(ctx, "k1")
		xt.ErrorIs(t, err, ErrNil)
		xt.Empty(t, value)

		err = client.SetWithTTL(ctx, "k1", "v1", time.Millisecond)
		xt.NoError(t, err)
		time.Sleep(2 * time.Millisecond)

		value, err = client.Get(ctx, "k1")
		xt.ErrorIs(t, err, ErrNil)
		xt.Empty(t, value)
	})

	t.Run("setnx", func(t *testing.T) {
		ok, err := client.SetNX(ctx, "k2", "v2", 0)
		xt.NoError(t, err)
		xt.True(t, ok)

		ok, err = client.SetNX(ctx, "k2", "v2-1", 0)
		xt.ErrorIs(t, err, ErrNil)
		xt.False(t, ok)
	})

	t.Run("setxx", func(t *testing.T) {
		ok, err := client.SetXX(ctx, "k3", "v3", 0)
		xt.ErrorIs(t, err, ErrNil)
		xt.False(t, ok)

		_, err = client.Get(ctx, "k3")
		xt.ErrorIs(t, err, ErrNil)

		err = client.Set(ctx, "k3", "v3")
		xt.NoError(t, err)

		ok, err = client.SetXX(ctx, "k3", "v3", 0)
		xt.NoError(t, err)
		xt.True(t, ok)
	})

	t.Run("incr_decr", func(t *testing.T) {
		err := client.Set(ctx, "incr-1", "v3")
		xt.NoError(t, err)

		num, err := client.Incr(ctx, "incr-1") // invalid type
		xt.Error(t, err)
		xt.Equal(t, 0, num)

		num, err = client.Incr(ctx, "incr-2")
		xt.NoError(t, err)
		xt.Equal(t, 1, num)

		num, err = client.IncrBy(ctx, "incr-2", 3)
		xt.NoError(t, err)
		xt.Equal(t, 4, num)

		num, err = client.Decr(ctx, "incr-2")
		xt.NoError(t, err)
		xt.Equal(t, 3, num)

		num, err = client.DecrBy(ctx, "incr-2", 2)
		xt.NoError(t, err)
		xt.Equal(t, 1, num)
	})

	t.Run("GetDel", func(t *testing.T) {
		err := client.Set(ctx, "GetDel-1", "v3")
		xt.NoError(t, err)

		val, err := client.GetDel(ctx, "GetDel-1")
		xt.NoError(t, err)
		xt.Equal(t, "v3", val)

		val, err = client.GetDel(ctx, "GetDel-1")
		xt.ErrorIs(t, err, ErrNil)
		xt.Equal(t, "", val)
	})

	t.Run("GetSet", func(t *testing.T) {
		val, err := client.GetSet(ctx, "k5", "v5")
		xt.ErrorIs(t, err, ErrNil)
		xt.Equal(t, "", val)

		val, err = client.GetSet(ctx, "k5", "v5-1")
		xt.NoError(t, err)
		xt.Equal(t, "v5", val)

		val, err = client.GetSet(ctx, "k5", "v5-2")
		xt.NoError(t, err)
		xt.Equal(t, "v5-1", val)
	})

	t.Run("getrange", func(t *testing.T) {
		val, err := client.GetRange(ctx, "k6", 0, 4)
		xt.NoError(t, err)
		xt.Equal(t, "", val)

		xt.NoError(t, client.Set(ctx, "k6", "hello-world"))

		val, err = client.GetRange(ctx, "k6", 0, 4)
		xt.NoError(t, err)
		xt.Equal(t, "hello", val)

		val, err = client.GetRange(ctx, "k6", 0, 100)
		xt.NoError(t, err)
		xt.Equal(t, "hello-world", val)
	})

	t.Run("echo", func(t *testing.T) {
		val, err := client.Echo(ctx, "hello")
		xt.NoError(t, err)
		xt.Equal(t, "hello", val)
	})
}

func testSetKeyString(t *testing.T, client *Client, key string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	_, err := client.Del(ctx, key)
	xt.NoError(t, err)
	err = client.Set(ctx, key, "str")
	xt.NoError(t, err)
}
