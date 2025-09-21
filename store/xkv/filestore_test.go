//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-20

package xkv_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/fsgo/fst"

	"github.com/xanygo/anygo/store/xkv"
)

func TestFileStorage(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "xkv_file")
	ff := &xkv.FileStorage{
		DataDir: dir,
	}
	testStorage(t, ff)
}

func testStorage(t *testing.T, ff xkv.Storage) {
	t.Run("string", func(t *testing.T) {
		ss1 := ff.String("hello")
		got1, err1 := ss1.Get(context.Background())
		fst.NoError(t, err1)
		fst.Equal(t, "", got1)
		fst.NoError(t, ss1.Set(context.Background(), "world"))
		got2, err2 := ss1.Get(context.Background())
		fst.NoError(t, err2)
		fst.Equal(t, "world", got2)
		fst.NoError(t, ff.Delete(context.Background(), "hello"))
	})

	t.Run("list", func(t *testing.T) {
		list := ff.List("list1")
		fst.NoError(t, list.RPush(context.Background(), "1"))
		fst.NoError(t, list.RPush(context.Background(), "2"))
		var values []string
		err1 := list.RRange(context.Background(), func(val string) bool {
			values = append(values, val)
			return true
		})
		fst.NoError(t, err1)
		fst.Equal(t, []string{"2", "1"}, values)
		values = nil
		err2 := list.LRange(context.Background(), func(val string) bool {
			values = append(values, val)
			return true
		})
		fst.NoError(t, err2)
		fst.Equal(t, []string{"1", "2"}, values)

		values = nil
		err3 := list.Range(context.Background(), func(val string) bool {
			values = append(values, val)
			return true
		})
		fst.NoError(t, err3)
		fst.Len(t, values, 2)
	})

	t.Run("Hash", func(t *testing.T) {
		hh := ff.Hash("hash1")
		fst.NoError(t, hh.HSet(context.Background(), "key1", "value1"))
		value1, found1, err1 := hh.HGet(context.Background(), "key1")
		fst.NoError(t, err1)
		fst.True(t, found1)
		fst.Equal(t, "value1", value1)

		value2, found2, err2 := hh.HGet(context.Background(), "key2")
		fst.NoError(t, err2)
		fst.False(t, found2)
		fst.Equal(t, "", value2)

		all, err4 := hh.HGetAll(context.Background())
		fst.NoError(t, err4)
		fst.Equal(t, map[string]string{"key1": "value1"}, all)

		fst.NoError(t, hh.HDel(context.Background(), "key1"))
		value3, found3, err3 := hh.HGet(context.Background(), "key2")
		fst.NoError(t, err3)
		fst.False(t, found3)
		fst.Equal(t, "", value3)
	})

	t.Run("Set", func(t *testing.T) {
		set := ff.Set("set1")
		fst.NoError(t, set.SAdd(context.Background(), "v1"))
		got1, err1 := set.SMembers(context.Background())
		fst.NoError(t, err1)
		fst.Equal(t, []string{"v1"}, got1)

		fst.NoError(t, set.SAdd(context.Background(), "v2"))
		got2, err2 := set.SMembers(context.Background())
		fst.NoError(t, err2)
		fst.Equal(t, []string{"v1", "v2"}, got2)

		fst.NoError(t, set.SRem(context.Background(), "v1"))
		got3, err3 := set.SMembers(context.Background())
		fst.NoError(t, err3)
		fst.Equal(t, []string{"v2"}, got3)
	})

	t.Run("ZSet", func(t *testing.T) {
		zset := ff.ZSet("zset1")
		fst.NoError(t, zset.ZAdd(context.Background(), 1, "m1"))
		got1, found1, err1 := zset.ZScore(context.Background(), "m1")
		fst.NoError(t, err1)
		fst.True(t, found1)
		fst.Equal(t, 1, got1)

		fst.NoError(t, zset.ZAdd(context.Background(), 2, "m2"))
		fst.NoError(t, zset.ZAdd(context.Background(), 1.5, "m3"))
		var members []string
		var scores []float64
		zset.ZRange(context.Background(), func(member string, score float64) bool {
			members = append(members, member)
			scores = append(scores, score)
			return true
		})
		fst.Equal(t, []string{"m1", "m3", "m2"}, members)
		fst.Equal(t, []float64{1, 1.5, 2}, scores)

		fst.NoError(t, zset.ZRem(context.Background(), "m2"))
		got2, found2, err2 := zset.ZScore(context.Background(), "m2")
		fst.NoError(t, err2)
		fst.False(t, found2)
		fst.Equal(t, 0, got2)
	})
}
