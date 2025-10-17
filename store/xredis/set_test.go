//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-17

package xredis

import (
	"context"
	"testing"
	"time"

	"github.com/fsgo/fst"

	"github.com/xanygo/anygo/internal/redistest"
)

func TestClient_Set(t *testing.T) {
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

	t.Run("SAdd", func(t *testing.T) {
		num, err := client.SAdd(ctx, "s1", "v1")
		fst.NoError(t, err)
		fst.Equal(t, 1, num)

		num, err = client.SCard(ctx, "s1")
		fst.NoError(t, err)
		fst.Equal(t, 1, num)

		members, err := client.SMembers(ctx, "s1")
		fst.NoError(t, err)
		fst.Equal(t, []string{"v1"}, members)

		num, err = client.SAdd(ctx, "s2", "v1", "v2")
		fst.NoError(t, err)
		fst.Equal(t, 2, num)

		num, err = client.SCard(ctx, "s2")
		fst.NoError(t, err)
		fst.Equal(t, 2, num)

		members, err = client.SMembers(ctx, "s2")
		fst.NoError(t, err)
		fst.Equal(t, []string{"v1", "v2"}, members)

		ok, err := client.SIsMember(ctx, "s2", "v1")
		fst.NoError(t, err)
		fst.True(t, ok)
	})

	t.Run("SCard", func(t *testing.T) {
		num, err := client.SCard(ctx, "s3-not-found")
		fst.NoError(t, err)
		fst.Equal(t, 0, num)

		members, err := client.SMembers(ctx, "s3-not-found")
		fst.NoError(t, err)
		fst.Empty(t, members)
	})

	t.Run("SIsMember", func(t *testing.T) {
		num, err := client.SAdd(ctx, "s4", "v1")
		fst.NoError(t, err)
		fst.Equal(t, 1, num)

		ok, err := client.SIsMember(ctx, "s4", "v1")
		fst.NoError(t, err)
		fst.True(t, ok)

		val, err := client.SMIsMember(ctx, "s4", "v1", "v100")
		fst.NoError(t, err)
		fst.Equal(t, 2, len(val))
		fst.Equal(t, []bool{true, false}, val)

		ok, err = client.SIsMember(ctx, "s4", "v100")
		fst.NoError(t, err)
		fst.False(t, ok)

		ok, err = client.SIsMember(ctx, "s4-not-found", "v100")
		fst.NoError(t, err)
		fst.False(t, ok)

		val, err = client.SMIsMember(ctx, "s4-not-found", "v1", "v100")
		fst.NoError(t, err)
		fst.Equal(t, 2, len(val))
		fst.Equal(t, []bool{false, false}, val)

		testSetKeyString(t, client, "s5")
		ok, err = client.SIsMember(ctx, "s5", "v100")
		fst.ErrorContains(t, err, "WRONGTYPE")
		fst.False(t, ok)
	})

	t.Run("SPop", func(t *testing.T) {
		num, err := client.SAdd(ctx, "s6", "v1")
		fst.NoError(t, err)
		fst.Equal(t, 1, num)

		val, ok, err := client.SPop(ctx, "s6")
		fst.NoError(t, err)
		fst.Equal(t, "v1", val)
		fst.True(t, ok)

		val, ok, err = client.SPop(ctx, "s6") // empty set
		fst.NoError(t, err)
		fst.Equal(t, "", val)
		fst.False(t, ok)
	})

	t.Run("SRandMember", func(t *testing.T) {
		num, err := client.SAdd(ctx, "s7", "v1")
		fst.NoError(t, err)
		fst.Equal(t, 1, num)

		vals, err := client.SRandMember(ctx, "s7", 2)
		fst.NoError(t, err)
		fst.Equal(t, []string{"v1"}, vals)

		vals, err = client.SRandMember(ctx, "s7-not-found", 2)
		fst.NoError(t, err)
		fst.Empty(t, vals)
	})

	t.Run("SRem", func(t *testing.T) {
		num, err := client.SAdd(ctx, "s8", "v1")
		fst.NoError(t, err)
		fst.Equal(t, 1, num)

		num, err = client.SRem(ctx, "s8", "v1")
		fst.NoError(t, err)
		fst.Equal(t, 1, num)

		num, err = client.SRem(ctx, "s8", "v1", "v2") // empty
		fst.NoError(t, err)
		fst.Equal(t, 0, num)
	})
}
