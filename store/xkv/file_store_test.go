//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-20

package xkv_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/xanygo/anygo/store/xkv"
	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xt"
)

func TestFileStorage(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "xkv_file")
	ff := &xkv.FileStore{
		DataDir: dir,
	}
	testStringStorage(t, ff)
}

func TestFileStorageCipher(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "xkv_file")
	ff := &xkv.FileStore{
		DataDir: dir,
	}
	aes := &xcodec.AesOFB{
		Key: "hello",
	}
	// 检查加密后，存储二进制内容不会报错
	coder := xcodec.CodecWithCipher(xcodec.JSON, aes)

	type user struct {
		Name string
	}

	u1 := user{
		Name: "hi 韩梅梅",
	}

	t.Run("list", func(t *testing.T) {
		store := xkv.AsList[user](ff, coder, "list1")

		num, err := store.RPush(context.Background(), u1)
		xt.NoError(t, err)
		xt.Equal(t, 1, num)

		var cnt int
		err = store.Range(context.Background(), func(val user) bool {
			xt.Equal(t, val, u1)
			cnt++
			return true
		})
		xt.NoError(t, err)
		xt.Equal(t, cnt, 1)
	})

	t.Run("hash", func(t *testing.T) {
		store := xkv.AsHash[user](ff, coder, "hash1")

		err := store.HSet(context.Background(), "f1", u1)
		xt.NoError(t, err)

		us, err := store.HGetAll(context.Background())
		xt.NoError(t, err)
		xt.NotEmpty(t, us)
	})

	t.Run("zset", func(t *testing.T) {
		store := xkv.AsZSet[user](ff, coder, "zset1")
		err := store.ZAdd(context.Background(), 1, u1)
		xt.NoError(t, err)
		var cnt int
		err = store.ZRange(context.Background(), func(member user, score float64) bool {
			xt.Equal(t, member.Name, u1.Name)
			xt.Equal(t, score, 1)
			cnt++
			return true
		})
		xt.NoError(t, err)
		xt.Equal(t, cnt, 1)
	})
}
