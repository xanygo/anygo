//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-14

package xredis

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/fsgo/fst"

	"github.com/xanygo/anygo/internal/redistest"
)

func TestClientHash(t *testing.T) {
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

	t.Run("HSet", func(t *testing.T) {
		num, err := client.HSet(ctx, "h1", "f1", "v1")
		fst.NoError(t, err)
		fst.Equal(t, 1, num)

		num, err = client.HSet(ctx, "h1", "f1", "v1")
		fst.NoError(t, err)
		fst.Equal(t, 0, num)

		testSetKeyString(t, client, "h1")

		num, err = client.HSet(ctx, "h1", "f1", "v1")
		fst.Error(t, err)
		fst.Equal(t, 0, num)
	})

	t.Run("HSetMap", func(t *testing.T) {
		num, err := client.HSetMap(ctx, "h2", map[string]string{"f1": "v1", "f2": "v2"})
		fst.NoError(t, err)
		fst.Equal(t, 2, num)
	})

	t.Run("HSetEX", func(t *testing.T) {
		data := map[string]string{"f1": "v1", "f2": "v2"}
		ok, err := client.HSetEX(ctx, "h3", "", -1, data)
		fst.NoError(t, err)
		fst.True(t, ok)

		ok, err = client.HSetEX(ctx, "h3", "FNX", -1, data)
		fst.NoError(t, err)
		fst.False(t, ok)

		testSetKeyString(t, client, "h3")

		ok, err = client.HSetEX(ctx, "h3", "", -1, data)
		fst.Error(t, err)
		fst.False(t, ok)
	})

	t.Run("HSetNX", func(t *testing.T) {
		ok, err := client.HSetNX(ctx, "h4", "f1", "v1")
		fst.NoError(t, err)
		fst.True(t, ok)

		ok, err = client.HSetNX(ctx, "h4", "f1", "v1")
		fst.NoError(t, err)
		fst.False(t, ok)

		testSetKeyString(t, client, "h4")

		ok, err = client.HSetNX(ctx, "h4", "f1", "v1")
		fst.Error(t, err)
		fst.False(t, ok)
	})

	t.Run("HStrLen", func(t *testing.T) {
		_, err := client.HSet(ctx, "h5", "f1", "v1")
		fst.NoError(t, err)

		num, err := client.HStrLen(ctx, "h5", "f1")
		fst.NoError(t, err)
		fst.Equal(t, 2, num)

		num, err = client.HDel(ctx, "h5", "f1", "f2")
		fst.NoError(t, err)
		fst.Equal(t, 1, num)
	})

	t.Run("HExists", func(t *testing.T) {
		ok, err := client.HExists(ctx, "h2", "f1")
		fst.NoError(t, err)
		fst.True(t, ok)
		testSetKeyString(t, client, "h2")

		ok, err = client.HExists(ctx, "h2", "f1")
		fst.Error(t, err)
		fst.False(t, ok)
	})

	t.Run("HGet", func(t *testing.T) {
		_, err := client.HSet(ctx, "h6", "f1", "v1")
		fst.NoError(t, err)

		val, err := client.HGet(ctx, "h6", "f1")
		fst.NoError(t, err)
		fst.Equal(t, "v1", val)

		val, err = client.HGet(ctx, "h6", "f2-not-exist")
		fst.ErrorIs(t, err, ErrNil)
		fst.Equal(t, "", val)
	})

	t.Run("HGetAll", func(t *testing.T) {
		data := map[string]string{"f1": "v1", "f2": "v2"}
		_, err := client.HSetMap(ctx, "h6", data)
		fst.NoError(t, err)

		val, err := client.HGetAll(ctx, "h6")
		fst.NoError(t, err)
		fst.Equal(t, data, val)

		val, err = client.HGetAll(ctx, "h6-not-exists")
		fst.NoError(t, err)
		fst.Empty(t, val)

		testSetKeyString(t, client, "h6")

		val, err = client.HGetAll(ctx, "h6")
		fst.Error(t, err)
		fst.Empty(t, val)
	})

	t.Run("HGetDel", func(t *testing.T) {
		data := map[string]string{"f1": "v1", "f2": "v2"}
		_, err := client.HSetMap(ctx, "h7", data)
		fst.NoError(t, err)

		val, err := client.HGetDel(ctx, "h7", "f1")
		fst.NoError(t, err)
		fst.Equal(t, map[string]string{"f1": "v1"}, val)

		val, err = client.HGetDel(ctx, "h7", "f1", "f2")
		fst.NoError(t, err)
		fst.Equal(t, map[string]string{"f2": "v2"}, val)
	})

	t.Run("HPersist", func(t *testing.T) {
		data := map[string]string{"f1": "v1", "f2": "v2"}
		_, err := client.HSetMap(ctx, "h7", data)
		fst.NoError(t, err)

		// todo
	})

	t.Run("HIncrBy", func(t *testing.T) {
		num, err := client.HIncrBy(ctx, "h8", "f1", 2)
		fst.NoError(t, err)
		fst.Equal(t, 2, num)

		testSetKeyString(t, client, "h8")
		num, err = client.HIncrBy(ctx, "h8", "f1", 2)
		fst.Error(t, err)
		fst.Equal(t, 0, num)
	})

	t.Run("HIncrFloat", func(t *testing.T) {
		num, err := client.HIncrFloat(ctx, "h9", "f1", 2)
		fst.NoError(t, err)
		fst.Equal(t, 2.0, num)

		keys, err := client.HKeys(ctx, "h9")
		fst.NoError(t, err)
		fst.Equal(t, []string{"f1"}, keys)

		num1, err := client.HLen(ctx, "h9")
		fst.NoError(t, err)
		fst.Equal(t, 1, num1)

		testSetKeyString(t, client, "h9")
		num, err = client.HIncrFloat(ctx, "h9", "f1", 2)
		fst.Error(t, err)
		fst.Equal(t, 0, num)
	})

	t.Run("HKeys", func(t *testing.T) {
		keys, err := client.HKeys(ctx, "h10-not-exists")
		fst.NoError(t, err)
		fst.Empty(t, keys)

		testSetKeyString(t, client, "h10")

		keys, err = client.HKeys(ctx, "h10")
		fst.Error(t, err)
		fst.Empty(t, keys)
	})

	t.Run("HMGet", func(t *testing.T) {
		data := map[string]string{"f1": "v1", "f2": "v2"}
		_, err := client.HSetMap(ctx, "h11", data)
		fst.NoError(t, err)

		val, err := client.HMGet(ctx, "h11", "f1", "f2", "f3")
		fst.NoError(t, err)
		fst.Equal(t, data, val)

		val, err = client.HMGet(ctx, "h11-not-exists", "f1", "f2", "f3")
		fst.NoError(t, err)
		fst.Empty(t, val)
	})

	t.Run("HVals", func(t *testing.T) {
		data := map[string]string{"f1": "v1", "f2": "v2"}
		_, err := client.HSetMap(ctx, "h12", data)
		fst.NoError(t, err)

		vals, err := client.HVals(ctx, "h12")
		fst.NoError(t, err)
		slices.Sort(vals)
		fst.Equal(t, []string{"v1", "v2"}, vals)
	})
}
