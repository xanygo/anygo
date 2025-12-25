//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-16

package xredis

import (
	"context"
	"testing"
	"time"

	"github.com/xanygo/anygo/internal/redistest"
	"github.com/xanygo/anygo/xt"
)

func TestClientList(t *testing.T) {
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

	t.Run("LRange", func(t *testing.T) {
		values, err := client.LRange(ctx, "l1", 0, -1)
		xt.NoError(t, err)
		xt.Empty(t, values)

		num, err := client.LPush(ctx, "l1", "v1")
		xt.NoError(t, err)
		xt.Equal(t, 1, num)

		values, err = client.LRange(ctx, "l1", 0, -1)
		xt.NoError(t, err)
		xt.Equal(t, []string{"v1"}, values)
	})

	t.Run("LPop", func(t *testing.T) {
		num, err := client.RPush(ctx, "k2", "v2")
		xt.NoError(t, err)
		xt.Equal(t, 1, num)

		value, err := client.LPop(ctx, "k2")
		xt.NoError(t, err)
		xt.Equal(t, "v2", value)

		value, err = client.LPop(ctx, "k2")
		xt.ErrorIs(t, err, ErrNil)
		xt.Equal(t, "", value)

		num, err = client.RPush(ctx, "k2", "v2", "v3", "v4")
		xt.NoError(t, err)
		xt.Equal(t, 3, num)
	})

	t.Run("LPopN", func(t *testing.T) {
		num, err := client.RPush(ctx, "k3", "v2", "v3", "v4")
		xt.NoError(t, err)
		xt.Equal(t, 3, num)
		values, err := client.LPopN(ctx, "k3", 2)
		xt.NoError(t, err)
		xt.Equal(t, []string{"v2", "v3"}, values)

		num, err = client.RPushX(ctx, "k3", "v5", "v6")
		xt.NoError(t, err)
		xt.Equal(t, 3, num)

		values, err = client.RPopN(ctx, "k3", 2)
		xt.NoError(t, err)
		xt.Equal(t, []string{"v6", "v5"}, values)
	})

	t.Run("RPop", func(t *testing.T) {
		got, err := client.RPop(ctx, "RPop-1")
		xt.ErrorIs(t, err, ErrNil)
		xt.Empty(t, got)

		num, err := client.RPush(ctx, "RPop-1", "v2", "v3", "v4")
		xt.NoError(t, err)
		xt.Equal(t, 3, num)

		got, err = client.RPop(ctx, "RPop-1")
		xt.NoError(t, err)
		xt.Equal(t, "v4", got)
	})

	t.Run("RPopN", func(t *testing.T) {
		got, err := client.RPopN(ctx, "RPopN-1", 2)
		xt.ErrorIs(t, err, ErrNil)
		xt.Empty(t, got)

		num, err := client.RPush(ctx, "RPopN-1", "v2", "v3", "v4")
		xt.NoError(t, err)
		xt.Equal(t, 3, num)

		got, err = client.RPopN(ctx, "RPopN-1", 2)
		xt.NoError(t, err)
		xt.Equal(t, []string{"v4", "v3"}, got)
	})

	t.Run("LLen", func(t *testing.T) {
		got, err := client.LLen(ctx, "LLen-1")
		xt.NoError(t, err)
		xt.Equal(t, 0, got)

		num, err := client.RPush(ctx, "LLen-1", "v2", "v3", "v4")
		xt.NoError(t, err)
		xt.Equal(t, 3, num)

		got, err = client.LLen(ctx, "LLen-1")
		xt.NoError(t, err)
		xt.Equal(t, 3, got)
	})
}
