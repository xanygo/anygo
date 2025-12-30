//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-12-27

package xredis

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/xanygo/anygo/internal/redistest"
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xt"
)

func TestClientTS(t *testing.T) {
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

	doTsAdd := func(t *testing.T, key string) {
		value := time.Now().UnixMilli()
		opt := &TSAddOption{
			ChunkSize:       48,
			DuplicatePolicy: TSDuplicateMax,
			Labels: map[string]string{
				"f1": "v1",
			},
		}
		got, err := client.TSAdd(ctx, key, value, 9, opt)
		xt.NoError(t, err)
		xt.Equal(t, value, got)
	}

	t.Run("TSAdd", func(t *testing.T) {
		doTsAdd(t, "TSAdd-1")

		opt1 := &TSAddOption{
			Retention:         1000,
			Encoding:          TSEncodingCompressed,
			ChunkSize:         48,
			DuplicatePolicy:   TSDuplicateMax,
			OnDuplicatePolicy: TSDuplicateFirst,
			Ignore: &TSIgnoreRule{
				MaxTimeDiff: 1,
				MaxValDiff:  9,
			},
			Labels: map[string]string{
				"f1": "v1",
				"f2": "v2",
			},
		}
		value := time.Now().UnixMilli()
		got, err := client.TSAdd(ctx, "TSAdd-2", value, 9, opt1)
		xt.NoError(t, err)
		xt.Equal(t, value, got)
	})

	t.Run("TSAlter", func(t *testing.T) {
		err := client.TSAlter(ctx, "TSAlter-1", &TSAlterOption{Retention: 1000})
		xt.Error(t, err)
		xt.True(t, xerror.IsNotFound(err))

		doTsAdd(t, "TSAlter-1")

		err = client.TSAlter(ctx, "TSAlter-1", &TSAlterOption{Retention: 1000})
		xt.NoError(t, err)
	})

	t.Run("TSCreate", func(t *testing.T) {
		err := client.TSCreate(ctx, "TSCreate-1", nil)
		xt.NoError(t, err)

		err = client.TSCreate(ctx, "TSCreate-1", nil) // 已存在
		xt.Error(t, err)
		xt.True(t, xerror.IsAlreadyExists(err))

		opt := &TTSCreateOption{
			Retention:       9,
			Encoding:        TSEncodingCompressed,
			ChunkSize:       48,
			DuplicatePolicy: TSDuplicateMax,
			Ignore: &TSIgnoreRule{
				MaxTimeDiff: 1,
				MaxValDiff:  2,
			},
			Labels: map[string]string{
				"f1": "v1",
			},
		}
		err = client.TSCreate(ctx, "TSCreate-2", opt)
		xt.NoError(t, err)
	})

	t.Run("TSCreateRule", func(t *testing.T) {
		err := client.TSCreateRule(ctx, "TSCreateRule-sk-1", "TSCreateRule-dk-1", "avg", 1, nil)
		xt.Error(t, err)
		xt.True(t, xerror.IsNotFound(err))

		doTsAdd(t, "TSCreateRule-sk-1")
		doTsAdd(t, "TSCreateRule-dk-1")

		err = client.TSCreateRule(ctx, "TSCreateRule-sk-1", "TSCreateRule-dk-1", "avg", 1, nil)
		xt.NoError(t, err)

		err = client.TSDelRule(ctx, "TSCreateRule-sk-1", "TSCreateRule-dk-1")
		xt.NoError(t, err)
	})

	t.Run("TSDecrBy", func(t *testing.T) {
		got, err := client.TSDecrBy(ctx, "TSDecrBy-1", 1.1, nil)
		xt.NoError(t, err)
		xt.Greater(t, got, 0)

		num := time.Now().UnixMilli()
		opt2 := &TSDecrByOption{
			TimeStamp:       num,
			Retention:       1,
			Encoding:        TSEncodingCompressed,
			ChunkSize:       48,
			DuplicatePolicy: TSDuplicateMax,
			Ignore: &TSIgnoreRule{
				MaxTimeDiff: 1,
				MaxValDiff:  9,
			},
			Labels: map[string]string{
				"f1": "v1",
			},
		}
		got, err = client.TSDecrBy(ctx, "TSDecrBy-2", 1.1, opt2)
		xt.NoError(t, err)
		xt.Equal(t, num, got)
	})

	t.Run("TSDel", func(t *testing.T) {
		num, err := client.TSDel(ctx, "TSDel-1", 0, time.Now().UnixMilli())
		xt.Error(t, err)
		xt.True(t, xerror.IsNotFound(err))
		xt.Equal(t, 0, num)

		doTsAdd(t, "TSDel-2")

		num, err = client.TSDel(ctx, "TSDel-2", 0, time.Now().UnixMilli())
		xt.NoError(t, err)
		xt.Equal(t, 1, num)
	})

	t.Run("TSDelRule", func(t *testing.T) {
		err := client.TSDelRule(ctx, "TSDelRule-sk-1", "TSDelRule-dk-1")
		xt.Error(t, err)
		xt.True(t, xerror.IsNotFound(err))
	})

	t.Run("TSGet", func(t *testing.T) {
		got, err := client.TSGet(ctx, "TSGet-1")
		xt.Error(t, err)
		xt.True(t, xerror.IsNotFound(err))
		xt.Nil(t, got)

		doTsAdd(t, "TSGet-2")

		got, err = client.TSGet(ctx, "TSGet-2")
		xt.NoError(t, err)
		xt.NotNil(t, got)
		xt.Equal(t, 9, got.Value)
		xt.Greater(t, got.Timestamp, 0)
	})

	t.Run("TSIncrBy", func(t *testing.T) {
		got, err := client.TSIncrBy(ctx, "TSIncrBy-1", 1.1, nil)
		xt.NoError(t, err)
		xt.Greater(t, got, 0)

		num := time.Now().UnixMilli()
		opt2 := &TSIncrByOption{
			TimeStamp:       num,
			Retention:       1,
			Encoding:        TSEncodingCompressed,
			ChunkSize:       48,
			DuplicatePolicy: TSDuplicateMax,
			Ignore: &TSIgnoreRule{
				MaxTimeDiff: 1,
				MaxValDiff:  9,
			},
			Labels: map[string]string{
				"f1": "v1",
			},
		}
		got, err = client.TSIncrBy(ctx, "TSIncrBy-2", 1.1, opt2)
		xt.NoError(t, err)
		xt.Equal(t, num, got)
	})

	t.Run("TSInfo", func(t *testing.T) {
		got, err := client.TSInfo(ctx, "TSInfo-1", false)
		xt.Error(t, err)
		xt.True(t, xerror.IsNotFound(err))
		xt.Empty(t, got)

		doTsAdd(t, "TSInfo-2")

		for i := 0; i < 100; i++ {
			doTsAdd(t, "TSInfo-3")
		}

		err = client.TSCreateRule(ctx, "TSInfo-2", "TSInfo-3", "avg", 1, nil)
		xt.NoError(t, err)

		got, err = client.TSInfo(ctx, "TSInfo-2", true)
		xt.NoError(t, err)
		t.Logf("TSInfo-2: %#v", got)
		xt.NotEmpty(t, got)
		xt.NotEmpty(t, got.Rules)
		xt.Equal(t, TSEncodingCompressed, got.ChunkType)
		xt.Equal(t, TSDuplicateMax, got.DuplicatePolicy)
		xt.Greater(t, got.MemoryUsage, 0)
		xt.Greater(t, got.TotalSamples, 0)
		xt.Greater(t, got.FirstTimestamp, 0)
		xt.Greater(t, got.LastTimestamp, 0)
		xt.Nil(t, got.SourceKey)

		got, err = client.TSInfo(ctx, "TSInfo-3", true)
		xt.NoError(t, err)
		xt.NotEmpty(t, got)
		t.Logf("TSInfo-3: %#v", got)
		xt.NotEmpty(t, got.Chunks)
		xt.Empty(t, got.Rules)
		xt.Equal(t, map[string]string{"f1": "v1"}, got.Labels)
		xt.Equal(t, 48, got.ChunkSize)
		xt.Equal(t, len(got.Chunks), int(got.ChunkCount))
		xt.NotNil(t, got.SourceKey)
	})

	t.Run("TSMGet", func(t *testing.T) {
		doTsAdd(t, "TSMGet-1")
		got, err := client.TSMGet(ctx, []string{"f1=v1"}, nil)
		xt.NoError(t, err)
		xt.NotEmpty(t, got)
		var item *TSSample
		for _, val := range got {
			if val.Key == "TSMGet-1" {
				item = val.Sample
			}
		}
		xt.NotEmpty(t, item)
		xt.NotEmpty(t, item.Timestamp)
		xt.NotEmpty(t, item.Timestamp)

		opt := &TSMGetOption{
			Latest:     true,
			WithLabels: true,
		}
		got, err = client.TSMGet(ctx, []string{"f1=v1"}, opt)
		xt.NoError(t, err)
		xt.NotEmpty(t, got)
	})

	t.Run("TSQueryIndex", func(t *testing.T) {
		doTsAdd(t, "TSQueryIndex-1")
		got, err := client.TSQueryIndex(ctx, "abc=not-found")
		xt.NoError(t, err)
		xt.Empty(t, got)

		got, err = client.TSQueryIndex(ctx, "f1=v1")
		xt.NoError(t, err)
		xt.NotEmpty(t, got)
		xt.SliceContains(t, got, "TSQueryIndex-1")
	})

	t.Run("TSRange", func(t *testing.T) {
		doTsAdd(t, "TSRange-1")
		got, err := client.TSRange(ctx, "TSRange-1", "-", "+", "avg", time.Now().Add(-1*time.Hour).UnixMilli(), nil)
		xt.NoError(t, err)
		xt.NotEmpty(t, got)
		for _, val := range got {
			xt.NotEmpty(t, val.Value)
			xt.NotEmpty(t, val.Timestamp)
		}

		opt := &TSRangeOption{
			Latest: true,
			Count:  100,
			Align:  "-",
			Empty:  true,
		}
		t1 := time.Now().Add(-2 * time.Hour).UnixMilli()
		t2 := time.Now().Add(2 * time.Hour).UnixMilli()
		got, err = client.TSRange(ctx, "TSRange-1", strconv.FormatInt(t1, 10), strconv.FormatInt(t2, 10), "avg", time.Now().Add(-1*time.Hour).UnixMilli(), opt)
		xt.NoError(t, err)
		xt.NotEmpty(t, got)
	})
	t.Run("TSRevRange", func(t *testing.T) {
		doTsAdd(t, "TSRevRange-1")
		got, err := client.TSRevRange(ctx, "TSRevRange-1", "-", "+", "avg", time.Now().Add(-1*time.Hour).UnixMilli(), nil)
		xt.NoError(t, err)
		xt.NotEmpty(t, got)
		for _, val := range got {
			xt.NotEmpty(t, val.Value)
			xt.NotEmpty(t, val.Timestamp)
		}
	})

	t.Run("TSMRange", func(t *testing.T) {
		got, err := client.TSMRange(ctx, "-", "+", "avg", 1, []string{"f1=v1"}, nil)
		xt.NoError(t, err)
		xt.NotEmpty(t, got)

		opt := &TSMRangeOption{
			Latest:     true,
			WithLabels: true,
			GroupBy:    "f1",
			Reduce:     "max",
		}
		got, err = client.TSMRange(ctx, "-", "+", "avg", 1, []string{"f1=v1"}, opt)
		xt.NoError(t, err)
		xt.NotEmpty(t, got)
		for _, val := range got {
			xt.NotEmpty(t, val.Key)
			xt.NotEmpty(t, val.Labels)
			xt.NotEmpty(t, val.Metadata)
			xt.NotEmpty(t, val.Sources)
			xt.NotEmpty(t, val.Samples)
		}
	})
}
