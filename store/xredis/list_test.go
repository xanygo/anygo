//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-16

package xredis

import (
	"context"
	"testing"
	"time"

	"github.com/fsgo/fst"

	"github.com/xanygo/anygo/internal/redistest"
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
	fst.NoError(t, errClient)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	t.Run("LRange", func(t *testing.T) {
		values, err := client.LRange(ctx, "l1", 0, -1)
		fst.NoError(t, err)
		fst.Empty(t, values)

		num, err := client.LPush(ctx, "l1", "v1")
		fst.NoError(t, err)
		fst.Equal(t, 1, num)

		values, err = client.LRange(ctx, "l1", 0, -1)
		fst.NoError(t, err)
		fst.Equal(t, []string{"v1"}, values)
	})

	t.Run("LPop", func(t *testing.T) {
		num, err := client.RPush(ctx, "k2", "v2")
		fst.NoError(t, err)
		fst.Equal(t, 1, num)

		value, err := client.LPop(ctx, "k2")
		fst.NoError(t, err)
		fst.Equal(t, "v2", value)

		value, err = client.LPop(ctx, "k2")
		fst.ErrorIs(t, err, ErrNil)
		fst.Equal(t, "", value)

		num, err = client.RPush(ctx, "k2", "v2", "v3", "v4")
		fst.NoError(t, err)
		fst.Equal(t, 3, num)
	})

	t.Run("LPopN", func(t *testing.T) {
		num, err := client.RPush(ctx, "k3", "v2", "v3", "v4")
		fst.NoError(t, err)
		fst.Equal(t, 3, num)
		values, err := client.LPopN(ctx, "k3", 2)
		fst.NoError(t, err)
		fst.Equal(t, []string{"v2", "v3"}, values)

		num, err = client.RPushX(ctx, "k3", "v5", "v6")
		fst.NoError(t, err)
		fst.Equal(t, 3, num)

		values, err = client.RPopN(ctx, "k3", 2)
		fst.NoError(t, err)
		fst.Equal(t, []string{"v6", "v5"}, values)
	})
}
