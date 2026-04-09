//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-17

package xredis

import (
	"context"
	"testing"
	"time"

	"github.com/xanygo/anygo/internal/redistest"
	"github.com/xanygo/anygo/xt"
)

func TestClient_Set(t *testing.T) {
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

	t.Run("SAdd", func(t *testing.T) {
		num, err := client.SAdd(ctx, "s1", "v1")
		xt.NoError(t, err)
		xt.Equal(t, num, 1)

		num, err = client.SCard(ctx, "s1")
		xt.NoError(t, err)
		xt.Equal(t, num, 1)

		members, err := client.SMembers(ctx, "s1")
		xt.NoError(t, err)
		xt.Equal(t, members, []string{"v1"})

		num, err = client.SAdd(ctx, "s2", "v1", "v2")
		xt.NoError(t, err)
		xt.Equal(t, num, 2)

		num, err = client.SCard(ctx, "s2")
		xt.NoError(t, err)
		xt.Equal(t, num, 2)

		members, err = client.SMembers(ctx, "s2")
		xt.NoError(t, err)
		xt.Equal(t, members, []string{"v1", "v2"})

		ok, err := client.SIsMember(ctx, "s2", "v1")
		xt.NoError(t, err)
		xt.True(t, ok)
	})

	t.Run("SCard", func(t *testing.T) {
		num, err := client.SCard(ctx, "s3-not-found")
		xt.NoError(t, err)
		xt.Equal(t, num, 0)

		members, err := client.SMembers(ctx, "s3-not-found")
		xt.NoError(t, err)
		xt.Empty(t, members)
	})

	t.Run("SIsMember", func(t *testing.T) {
		num, err := client.SAdd(ctx, "s4", "v1")
		xt.NoError(t, err)
		xt.Equal(t, num, 1)

		ok, err := client.SIsMember(ctx, "s4", "v1")
		xt.NoError(t, err)
		xt.True(t, ok)

		val, err := client.SMIsMember(ctx, "s4", "v1", "v100")
		xt.NoError(t, err)
		xt.Equal(t, len(val), 2)
		xt.Equal(t, val, []bool{true, false})

		ok, err = client.SIsMember(ctx, "s4", "v100")
		xt.NoError(t, err)
		xt.False(t, ok)

		ok, err = client.SIsMember(ctx, "s4-not-found", "v100")
		xt.NoError(t, err)
		xt.False(t, ok)

		val, err = client.SMIsMember(ctx, "s4-not-found", "v1", "v100")
		xt.NoError(t, err)
		xt.Equal(t, len(val), 2)
		xt.Equal(t, val, []bool{false, false})

		testSetKeyString(t, client, "s5")
		ok, err = client.SIsMember(ctx, "s5", "v100")
		xt.ErrorContains(t, err, "WRONGTYPE")
		xt.False(t, ok)
	})

	t.Run("SPop", func(t *testing.T) {
		num, err := client.SAdd(ctx, "s6", "v1")
		xt.NoError(t, err)
		xt.Equal(t, num, 1)

		val, ok, err := client.SPop(ctx, "s6")
		xt.NoError(t, err)
		xt.Equal(t, val, "v1")
		xt.True(t, ok)

		val, ok, err = client.SPop(ctx, "s6") // empty set
		xt.NoError(t, err)
		xt.Equal(t, val, "")
		xt.False(t, ok)
	})

	t.Run("SRandMember", func(t *testing.T) {
		num, err := client.SAdd(ctx, "s7", "v1")
		xt.NoError(t, err)
		xt.Equal(t, num, 1)

		vals, err := client.SRandMember(ctx, "s7", 2)
		xt.NoError(t, err)
		xt.Equal(t, vals, []string{"v1"})

		vals, err = client.SRandMember(ctx, "s7-not-found", 2)
		xt.NoError(t, err)
		xt.Empty(t, vals)
	})

	t.Run("SRem", func(t *testing.T) {
		num, err := client.SAdd(ctx, "s8", "v1")
		xt.NoError(t, err)
		xt.Equal(t, num, 1)

		num, err = client.SRem(ctx, "s8", "v1")
		xt.NoError(t, err)
		xt.Equal(t, num, 1)

		num, err = client.SRem(ctx, "s8", "v1", "v2") // empty
		xt.NoError(t, err)
		xt.Equal(t, num, 0)
	})

	t.Run("SUnion", func(t *testing.T) {
		got, err := client.SUnion(ctx, "s-u-1")
		xt.NoError(t, err)
		xt.Empty(t, got)

		num, err := client.SAdd(ctx, "s-u-1", "v1", "v2")
		xt.NoError(t, err)
		xt.Equal(t, num, 2)

		got, err = client.SUnion(ctx, "s-u-1")
		xt.NoError(t, err)
		xt.SliceSortEqual(t, []string{"v1", "v2"}, got)

		got, err = client.SUnion(ctx, "s-u-1", "s-u-not-found")
		xt.NoError(t, err)
		xt.SliceSortEqual(t, []string{"v1", "v2"}, got)

		num, err = client.SAdd(ctx, "s-u-2", "v3", "v2")
		xt.NoError(t, err)
		xt.Equal(t, num, 2)

		got, err = client.SUnion(ctx, "s-u-1", "s-u-not-found", "s-u-2")
		xt.NoError(t, err)
		xt.SliceSortEqual(t, []string{"v1", "v2", "v3"}, got)
	})

	t.Run("SUnionStore", func(t *testing.T) {
		got, err := client.SUnionStore(ctx, "SUnionStore-dest-1", "s-u-s-1")
		xt.NoError(t, err)
		xt.Equal(t, got, 0)

		num, err := client.SAdd(ctx, "s-u-s-1", "v1", "v2")
		xt.NoError(t, err)
		xt.Equal(t, num, 2)

		got, err = client.SUnionStore(ctx, "SUnionStore-dest-1", "s-u-s-1")
		xt.NoError(t, err)
		xt.Equal(t, got, 2)
	})
}
