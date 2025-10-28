//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-24

package xkv_test

import (
	"context"
	"testing"

	"github.com/xanygo/anygo/store/xkv"
	"github.com/xanygo/anygo/xt"
)

func testStorage(t *testing.T, ff xkv.StringStorage) {
	t.Run("string", func(t *testing.T) {
		ss1 := ff.String("hello")
		got1, found1, err1 := ss1.Get(context.Background())
		xt.NoError(t, err1)
		xt.False(t, found1)
		xt.Equal(t, "", got1)
		xt.NoError(t, ss1.Set(context.Background(), "world"))
		got2, found2, err2 := ss1.Get(context.Background())
		xt.True(t, found2)
		xt.NoError(t, err2)
		xt.Equal(t, "world", got2)

		got3, err3 := ff.Has(context.Background(), "hello")
		xt.NoError(t, err3)
		xt.True(t, got3)

		xt.NoError(t, ff.Delete(context.Background(), "hello"))
	})

	t.Run("list", func(t *testing.T) {
		list := ff.List("list1")
		_, err1 := list.RPush(context.Background(), "1")
		xt.NoError(t, err1)

		_, err2 := list.RPush(context.Background(), "2")
		xt.NoError(t, err2)
		var values []string
		err3 := list.RRange(context.Background(), func(val string) bool {
			values = append(values, val)
			return true
		})
		xt.NoError(t, err3)
		xt.Equal(t, []string{"2", "1"}, values)

		values = nil
		err4 := list.LRange(context.Background(), func(val string) bool {
			values = append(values, val)
			return true
		})
		xt.NoError(t, err4)
		xt.Equal(t, []string{"1", "2"}, values)

		values = nil
		err5 := list.Range(context.Background(), func(val string) bool {
			values = append(values, val)
			return true
		})
		xt.NoError(t, err5)
		xt.Len(t, values, 2)
	})

	t.Run("Hash", func(t *testing.T) {
		hh := ff.Hash("hash1")
		xt.NoError(t, hh.HSet(context.Background(), "key1", "value1"))
		value1, found1, err1 := hh.HGet(context.Background(), "key1")
		xt.NoError(t, err1)
		xt.True(t, found1)
		xt.Equal(t, "value1", value1)

		value2, found2, err2 := hh.HGet(context.Background(), "key2")
		xt.NoError(t, err2)
		xt.False(t, found2)
		xt.Equal(t, "", value2)

		all, err4 := hh.HGetAll(context.Background())
		xt.NoError(t, err4)
		xt.Equal(t, map[string]string{"key1": "value1"}, all)

		xt.NoError(t, hh.HDel(context.Background(), "key1"))
		value3, found3, err3 := hh.HGet(context.Background(), "key2")
		xt.NoError(t, err3)
		xt.False(t, found3)
		xt.Equal(t, "", value3)
	})

	t.Run("Set", func(t *testing.T) {
		set := ff.Set("set1")
		_, err1 := set.SAdd(context.Background(), "v1")
		xt.NoError(t, err1)

		got1, err2 := set.SMembers(context.Background())
		xt.NoError(t, err2)
		xt.Equal(t, []string{"v1"}, got1)

		_, err3 := set.SAdd(context.Background(), "v2")
		xt.NoError(t, err3)

		got2, err4 := set.SMembers(context.Background())
		xt.NoError(t, err4)
		xt.Equal(t, []string{"v1", "v2"}, got2)

		xt.NoError(t, set.SRem(context.Background(), "v1"))
		got3, err3 := set.SMembers(context.Background())
		xt.NoError(t, err3)
		xt.Equal(t, []string{"v2"}, got3)
	})

	t.Run("ZSet", func(t *testing.T) {
		zset := ff.ZSet("zset1")
		xt.NoError(t, zset.ZAdd(context.Background(), 1, "m1"))
		got1, found1, err1 := zset.ZScore(context.Background(), "m1")
		xt.NoError(t, err1)
		xt.True(t, found1)
		xt.Equal(t, 1, got1)

		xt.NoError(t, zset.ZAdd(context.Background(), 2, "m2"))
		xt.NoError(t, zset.ZAdd(context.Background(), 1.5, "m3"))
		var members []string
		var scores []float64
		zset.ZRange(context.Background(), func(member string, score float64) bool {
			members = append(members, member)
			scores = append(scores, score)
			return true
		})
		xt.Equal(t, []string{"m1", "m3", "m2"}, members)
		xt.Equal(t, []float64{1, 1.5, 2}, scores)

		xt.NoError(t, zset.ZRem(context.Background(), "m2"))
		got2, found2, err2 := zset.ZScore(context.Background(), "m2")
		xt.NoError(t, err2)
		xt.False(t, found2)
		xt.Equal(t, 0, got2)
	})
}

func benchStorage(b *testing.B, st xkv.StringStorage) {
	b.Run("string", func(b *testing.B) {
		s1 := st.String("str1")
		b.Run("set", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = s1.Set(context.Background(), "v1")
			}
		})
		b.Run("get", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				s1.Get(context.Background())
			}
		})
	})

	b.Run("list", func(b *testing.B) {
		l1 := st.List("list1")
		for i := 0; i < b.N; i++ {
			_, err1 := l1.LPush(context.Background(), "v1")
			if err1 != nil {
				b.Fatal(err1.Error())
			}
			got, found, err2 := l1.LPop(context.Background())
			if err2 != nil {
				b.Fatal(err2.Error())
			}
			if !found || got != "v1" {
				b.Fatalf("not found or value is wrong %v %v", found, got)
			}
		}
	})
}
