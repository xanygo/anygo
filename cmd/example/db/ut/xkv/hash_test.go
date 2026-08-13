package xkv

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/xanygo/anygo/store/xkv"
	"github.com/xanygo/anygo/xt"
)

func checkHash(t *testing.T, kvs xkv.StringStorage) {
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()

	t.Run("hash1", func(t *testing.T) {
		ha := kvs.Hash("hash1")
		value, found, err := ha.HGet(ctx, "f1")
		xt.NoError(t, err)
		xt.False(t, found)
		xt.Equal(t, value, "")

		err = ha.HSet(ctx, "f1", "v1")
		xt.NoError(t, err)

		value, found, err = ha.HGet(ctx, "f1")
		xt.NoError(t, err)
		xt.True(t, found)
		xt.Equal(t, value, "v1")

		err = ha.HSet(ctx, "f1", "v2")
		xt.NoError(t, err)

		value, found, err = ha.HGet(ctx, "f1")
		xt.NoError(t, err)
		xt.True(t, found)
		xt.Equal(t, value, "v2")
	})

	checkGetHa := func(t *testing.T, ha xkv.Hash[string], field string, want string) {
		t.Helper()
		value, found, err1 := ha.HGet(ctx, field)
		xt.NoError(t, err1)
		xt.Equal(t, value, want)
		xt.Equal(t, found, want != "")

		found2, err2 := ha.HExists(ctx, field)
		xt.NoError(t, err2)
		xt.Equal(t, found2, want != "")
	}

	t.Run("hash2", func(t *testing.T) {
		ha := kvs.Hash("hash2")
		err := ha.HDel(ctx, "f1")
		xt.NoError(t, err)

		checkGet := func(t *testing.T, field string, want string) {
			checkGetHa(t, ha, field, want)
		}

		vs := map[string]string{"f1": "v1", "f2": "v2"}
		err = ha.HMSet(ctx, vs)
		xt.NoError(t, err)

		checkGet(t, "f1", "v1")
		checkGet(t, "f2", "v2")
		checkGet(t, "f3", "")

		all := map[string]string{}
		err = ha.HRange(ctx, func(field string, value string) bool {
			all[field] = value
			return true
		})
		xt.NoError(t, err)
		xt.Equal(t, all, vs)

		got, err := ha.HGetAll(ctx)
		xt.NoError(t, err)
		xt.Equal(t, got, vs)

		err = ha.HDel(ctx, "f1", "f3")
		xt.NoError(t, err)
		checkGet(t, "f1", "")
		checkGet(t, "f2", "v2")
		checkGet(t, "f3", "")

		has, err := kvs.Has(ctx, "hash2")
		xt.NoError(t, err)
		xt.True(t, has)
	})

	t.Run("hash3-HIncrBy", func(t *testing.T) {
		ha := kvs.Hash("t2-hash3")
		num, err := ha.HIncrBy(ctx, "f1", 1)
		xt.NoError(t, err)
		xt.Equal(t, num, 1)
		checkGetHa(t, ha, "f1", "1")

		num, err = ha.HIncrBy(ctx, "f1", 3)
		xt.NoError(t, err)
		xt.Equal(t, num, 4)
		checkGetHa(t, ha, "f1", "4")
	})

	t.Run("hash4-hlen", func(t *testing.T) {
		ha := kvs.Hash("t2-hash4")
		num, err := ha.HLen(ctx)
		xt.NoError(t, err)
		xt.Equal(t, num, 0)

		for i := 0; i < 10; i++ {
			num, err = ha.HIncrBy(ctx, fmt.Sprintf("f-%d", i), int64(i))
			xt.NoError(t, err)
			xt.Equal(t, num, int64(i))

			num, err = ha.HLen(ctx)
			xt.NoError(t, err)
			xt.Equal(t, num, int64(i)+1)
		}
	})

	t.Run("hmget1", func(t *testing.T) {
		ha := kvs.Hash("t2-hmget1")
		result, err := ha.HMGet(ctx, "f1", "f2")
		xt.NoError(t, err)
		xt.Empty(t, result)
		kv1 := map[string]string{
			"f1": "v1",
			"f2": "v2",
			"f3": "v3",
		}
		err = ha.HMSet(ctx, kv1)
		xt.NoError(t, err)

		result, err = ha.HMGet(ctx, "f1", "f2")
		xt.NoError(t, err)
		xt.Equal(t, result, map[string]string{"f1": "v1", "f2": "v2"})
	})
}
