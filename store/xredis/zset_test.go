//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-16

package xredis

import (
	"context"
	"maps"
	"math"
	"testing"
	"time"

	"github.com/xanygo/anygo/internal/redistest"
	"github.com/xanygo/anygo/xt"
)

func TestClientZSet(t *testing.T) {
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

	t.Run("ZAdd", func(t *testing.T) {
		ok, err := client.ZAdd(ctx, "z1", 1, "f1")
		xt.NoError(t, err)
		xt.True(t, ok)

		ok, err = client.ZAdd(ctx, "z1", 1, "f1")
		xt.NoError(t, err)
		xt.False(t, ok)

		ok, err = client.ZAdd(ctx, "z1", 2, "f1")
		xt.NoError(t, err)
		xt.False(t, ok)

		ok, err = client.ZAdd(ctx, "z1", math.Inf(-1), "f1")
		xt.NoError(t, err)
		xt.False(t, ok)
	})

	t.Run("ZAddIncr", func(t *testing.T) {
		num, err := client.ZAddIncr(ctx, "ZAddIncr-1", 1, "f1")
		xt.NoError(t, err)
		xt.Equal(t, 1, num)

		num, err = client.ZAddIncr(ctx, "ZAddIncr-1", 1.1, "f1")
		xt.NoError(t, err)
		xt.Equal(t, 2.1, num)

		cn, err := client.ZCard(ctx, "ZAddIncr-1")
		xt.NoError(t, err)
		xt.Equal(t, int64(1), cn)
	})

	t.Run("ZAddOpt", func(t *testing.T) {
		ok, err := client.ZAdd(ctx, "z2", 1, "f1")
		xt.NoError(t, err)
		xt.True(t, ok)

		num, err := client.ZAddOpt(ctx, "z2", []string{"NX"}, 1, "f1")
		xt.NoError(t, err)
		xt.Equal(t, 0, num)
	})

	data := map[string]float64{
		"f1": 1,
		"f2": 2,
	}
	t.Run("ZAddMap", func(t *testing.T) {
		data := maps.Clone(data)
		num, err := client.ZAddMap(ctx, "z3", data)
		xt.NoError(t, err)
		xt.Equal(t, 2, num)

		data["f1"] = 3
		num, err = client.ZAddMap(ctx, "z3", data)
		xt.NoError(t, err)
		xt.Equal(t, 0, num)
	})

	t.Run("ZAddMapOpt", func(t *testing.T) {
		num, err := client.ZAddMapOpt(ctx, "z4", []string{"NX"}, data)
		xt.NoError(t, err)
		xt.Equal(t, 2, num)

		num, err = client.ZAddMapOpt(ctx, "z4", []string{"NX"}, data)
		xt.NoError(t, err)
		xt.Equal(t, 0, num)

		num, err = client.ZCard(ctx, "z4")
		xt.NoError(t, err)
		xt.Equal(t, 2, num)
	})

	t.Run("ZCount", func(t *testing.T) {
		num, err := client.ZCount(ctx, "ZCount-1", 0, math.Inf(1))
		xt.NoError(t, err)
		xt.Equal(t, 0, num)

		ok, err := client.ZAdd(ctx, "ZCount-1", 1, "f1")
		xt.NoError(t, err)
		xt.True(t, ok)

		num, err = client.ZCount(ctx, "ZCount-1", 0, math.Inf(1))
		xt.NoError(t, err)
		xt.Equal(t, 1, num)
	})

	t.Run("ZDiff", func(t *testing.T) {
		ok, err := client.ZAdd(ctx, "ZDiff-1", 1, "f1")
		xt.NoError(t, err)
		xt.True(t, ok)

		got, err := client.ZDiff(ctx, "ZDiff-1")
		xt.Error(t, err)
		xt.Empty(t, got)

		got, err = client.ZDiff(ctx, "ZDiff-1", "ZDiff-2")
		xt.NoError(t, err)
		xt.Equal(t, []string{"f1"}, got)
	})

	t.Run("ZDiffWithScores", func(t *testing.T) {
		ok, err := client.ZAdd(ctx, "ZDiffWithScores-1", 1, "f1")
		xt.NoError(t, err)
		xt.True(t, ok)

		got, err := client.ZDiffWithScores(ctx, "ZDiffWithScores-1")
		xt.Error(t, err)
		xt.Empty(t, got)

		got, err = client.ZDiffWithScores(ctx, "ZDiffWithScores-1", "ZDiffWithScores-2")
		xt.NoError(t, err)
		xt.Equal(t, []Z{{Member: "f1", Score: 1}}, got)
	})

	t.Run("ZDiffStore", func(t *testing.T) {
		got, err := client.ZDiffStore(ctx, "ZDiffStore-dest-1", "ZDiffStore-1")
		xt.NoError(t, err)
		xt.Equal(t, 0, got)

		ok, err := client.ZAdd(ctx, "ZDiffStore-1", 1, "f1")
		xt.NoError(t, err)
		xt.True(t, ok)

		got, err = client.ZDiffStore(ctx, "ZDiffStore-dest-1", "ZDiffStore-1")
		xt.NoError(t, err)
		xt.Equal(t, 1, got)

		got, err = client.ZDiffStore(ctx, "ZDiffStore-dest-1", "ZDiffStore-1", "ZDiffStore-2")
		xt.NoError(t, err)
		xt.Equal(t, 1, got)
	})

	t.Run("ZIncrBy", func(t *testing.T) {
		got, err := client.ZIncrBy(ctx, "ZIncrBy-1", 1, "m1")
		xt.NoError(t, err)
		xt.Equal(t, 1, got)

		got, err = client.ZIncrBy(ctx, "ZIncrBy-1", 2.1, "m1")
		xt.NoError(t, err)
		xt.Equal(t, 3.1, got)
	})

	t.Run("ZInter", func(t *testing.T) {
		got, err := client.ZInter(ctx, "ZInter-1")
		xt.NoError(t, err)
		xt.Empty(t, got)

		num, err := client.ZAddMap(ctx, "ZInter-1", map[string]float64{"f1": 1, "f2": 2})
		xt.NoError(t, err)
		xt.Equal(t, 2, num)

		got, err = client.ZInter(ctx, "ZInter-1")
		xt.NoError(t, err)
		xt.SliceSortEqual(t, []string{"f1", "f2"}, got)

		got, err = client.ZInter(ctx, "ZInter-1", "ZInter-2")
		xt.NoError(t, err)
		xt.Empty(t, got)

		got1, err := client.ZInterWithScores(ctx, "ZInter-1")
		xt.NoError(t, err)
		xt.Equal(t, []Z{{Member: "f1", Score: 1}, {Member: "f2", Score: 2}}, got1)
	})

	t.Run("ZInterStore", func(t *testing.T) {
		got, err := client.ZInterStore(ctx, "ZInterStore-dest-1", "ZInterStore-1")
		xt.NoError(t, err)
		xt.Equal(t, 0, got)

		num, err := client.ZAddMap(ctx, "ZInterStore-1", map[string]float64{"f1": 1, "f2": 2})
		xt.NoError(t, err)
		xt.Equal(t, 2, num)

		got, err = client.ZInterStore(ctx, "ZInterStore-dest-1", "ZInterStore-1")
		xt.NoError(t, err)
		xt.Equal(t, 2, got)

		// 目标已存在，则覆盖
		got, err = client.ZInterStore(ctx, "ZInterStore-dest-1", "ZInterStore-1")
		xt.NoError(t, err)
		xt.Equal(t, 2, got)
	})

	t.Run("ZLexCount", func(t *testing.T) {
		got, err := client.ZLexCount(ctx, "ZLexCount-1", "-", "+")
		xt.NoError(t, err)
		xt.Equal(t, 0, got)

		num, err := client.ZAddMap(ctx, "ZLexCount-1", map[string]float64{"f1": 1, "f2": 2})
		xt.NoError(t, err)
		xt.Equal(t, 2, num)

		got, err = client.ZLexCount(ctx, "ZLexCount-1", "-", "+")
		xt.NoError(t, err)
		xt.Equal(t, 2, got)
	})

	t.Run("ZMPop", func(t *testing.T) {
		gotFrom, gotMem, err := client.ZMPop(ctx, "ZMPop-1", nil, true, 0)
		xt.ErrorIs(t, err, ErrNil)
		xt.Empty(t, gotFrom)
		xt.Empty(t, gotMem)

		num, err := client.ZAddMap(ctx, "ZMPop-1", map[string]float64{"f1": 1, "f2": 2})
		xt.NoError(t, err)
		xt.Equal(t, 2, num)

		gotFrom, gotMem, err = client.ZMPop(ctx, "ZMPop-1", nil, true, 2)
		xt.NoError(t, err)
		xt.Equal(t, "ZMPop-1", gotFrom)
		xt.Equal(t, []Z{{Member: "f1", Score: 1}, {Member: "f2", Score: 2}}, gotMem)
	})

	t.Run("ZMScore", func(t *testing.T) {
		got, err := client.ZMScore(ctx, "ZMScore-1", "m1", "m2")
		xt.NoError(t, err)
		xt.Empty(t, got)

		num, err := client.ZAddMap(ctx, "ZMScore-1", map[string]float64{"m1": 1, "m2": 2, "m3": 3})
		xt.NoError(t, err)
		xt.Equal(t, 3, num)

		got, err = client.ZMScore(ctx, "ZMScore-1", "m1", "m2", "m8")
		xt.NoError(t, err)
		xt.Equal(t, map[string]float64{"m1": 1, "m2": 2}, got)
	})

	t.Run("ZPopMax", func(t *testing.T) {
		got, err := client.ZPopMax(ctx, "ZPopMax-1", 2)
		xt.NoError(t, err)
		xt.Empty(t, got)

		num, err := client.ZAddMap(ctx, "ZPopMax-1", map[string]float64{"m1": 1, "m2": 2, "m3": 3})
		xt.NoError(t, err)
		xt.Equal(t, 3, num)

		got, err = client.ZPopMax(ctx, "ZPopMax-1", 2)
		xt.NoError(t, err)
		xt.Equal(t, []Z{{Member: "m3", Score: 3}, {Member: "m2", Score: 2}}, got)
	})

	t.Run("ZPopMin", func(t *testing.T) {
		got, err := client.ZPopMin(ctx, "ZPopMin-1", 2)
		xt.NoError(t, err)
		xt.Empty(t, got)

		num, err := client.ZAddMap(ctx, "ZPopMin-1", map[string]float64{"m1": 1, "m2": 2, "m3": 3})
		xt.NoError(t, err)
		xt.Equal(t, 3, num)

		got, err = client.ZPopMin(ctx, "ZPopMin-1", 2)
		xt.NoError(t, err)
		xt.Equal(t, []Z{{Member: "m1", Score: 1}, {Member: "m2", Score: 2}}, got)
	})

	t.Run("ZRandMember", func(t *testing.T) {
		got, ok, err := client.ZRandMember(ctx, "ZRandMember-1")
		xt.NoError(t, err)
		xt.Empty(t, got)
		xt.False(t, ok)

		num, err := client.ZAddMap(ctx, "ZRandMember-1", map[string]float64{"m1": 1, "m2": 2, "m3": 3})
		xt.NoError(t, err)
		xt.Equal(t, 3, num)

		got, ok, err = client.ZRandMember(ctx, "ZRandMember-1")
		xt.NoError(t, err)
		xt.True(t, ok)
		xt.SliceContains(t, []string{"m1", "m2", "m3"}, got)
	})

	t.Run("ZRandMemberN", func(t *testing.T) {
		got, err := client.ZRandMemberN(ctx, "ZRandMemberN-1", 2)
		xt.NoError(t, err)
		xt.Empty(t, got)

		num, err := client.ZAddMap(ctx, "ZRandMemberN-1", map[string]float64{"m1": 1, "m2": 2, "m3": 3})
		xt.NoError(t, err)
		xt.Equal(t, 3, num)

		got, err = client.ZRandMemberN(ctx, "ZRandMemberN-1", 2)
		xt.NoError(t, err)
		xt.Len(t, got, 2)
		xt.SliceContains(t, []string{"m1", "m2", "m3"}, got...)
	})

	t.Run("ZRandMemberWithScores", func(t *testing.T) {
		got, err := client.ZRandMemberWithScores(ctx, "ZRandMemberWithScores-1", 2)
		xt.NoError(t, err)
		xt.Empty(t, got)

		num, err := client.ZAddMap(ctx, "ZRandMemberWithScores-1", map[string]float64{"m1": 1, "m2": 2, "m3": 3})
		xt.NoError(t, err)
		xt.Equal(t, 3, num)

		got, err = client.ZRandMemberWithScores(ctx, "ZRandMemberWithScores-1", 2)
		xt.NoError(t, err)
		xt.Len(t, got, 2)
		xt.SliceContains(t, []string{"m1", "m2", "m3"}, got[0].Member, got[1].Member)
	})

	t.Run("ZRange", func(t *testing.T) {
		got, err := client.ZRange(ctx, "ZRange-1", ZRangeBy{Start: "0", Stop: "-1"})
		xt.NoError(t, err)
		xt.Empty(t, got)

		num, err := client.ZAddMap(ctx, "ZRange-1", map[string]float64{"m1": 1, "m2": 2, "m3": 3})
		xt.NoError(t, err)
		xt.Equal(t, 3, num)

		got, err = client.ZRange(ctx, "ZRange-1", ZRangeBy{Start: "0", Stop: "-1"})
		xt.NoError(t, err)
		xt.Equal(t, []string{"m1", "m2", "m3"}, got)

		got1, err := client.ZRangeWithScore(ctx, "ZRange-1", ZRangeBy{Start: "0", Stop: "-1"})
		xt.NoError(t, err)
		want1 := []Z{
			{Member: "m1", Score: 1},
			{Member: "m2", Score: 2},
			{Member: "m3", Score: 3},
		}
		xt.Equal(t, want1, got1)
	})

	t.Run("ZRank", func(t *testing.T) {
		got, err := client.ZRank(ctx, "ZRank-1", "m1")
		xt.ErrorIs(t, err, ErrNil)
		xt.Equal(t, 0, got)

		num, err := client.ZAddMap(ctx, "ZRank-1", map[string]float64{"m1": 1, "m2": 2, "m3": 3})
		xt.NoError(t, err)
		xt.Equal(t, 3, num)

		got, err = client.ZRank(ctx, "ZRank-1", "m1")
		xt.NoError(t, err)
		xt.Equal(t, 0, got)

		got, err = client.ZRank(ctx, "ZRank-1", "m2")
		xt.NoError(t, err)
		xt.Equal(t, 1, got)

		got, err = client.ZRank(ctx, "ZRank-1", "m1000-not-found")
		xt.ErrorIs(t, err, ErrNil)
		xt.Equal(t, 0, got)

		got, err = client.ZRevRank(ctx, "ZRank-1", "m1")
		xt.NoError(t, err)
		xt.Equal(t, 2, got)

		rank, score, err := client.ZRankWithScore(ctx, "ZRank-1", "m1")
		xt.NoError(t, err)
		xt.Equal(t, 0, rank)
		xt.Equal(t, 1, score)

		rank, score, err = client.ZRevRankWithScore(ctx, "ZRank-1", "m1")
		xt.NoError(t, err)
		xt.Equal(t, 2, rank)
		xt.Equal(t, 1, score)
	})

	t.Run("ZRem", func(t *testing.T) {
		got, err := client.ZRem(ctx, "ZRem-1", "m1")
		xt.NoError(t, err)
		xt.Equal(t, 0, got)

		num, err := client.ZAddMap(ctx, "ZRem-1", map[string]float64{"m1": 1, "m2": 2, "m3": 3})
		xt.NoError(t, err)
		xt.Equal(t, 3, num)

		got, err = client.ZRem(ctx, "ZRem-1", "m1", "m1000")
		xt.NoError(t, err)
		xt.Equal(t, 1, got)
	})

	t.Run("ZRemRangeByLex", func(t *testing.T) {
		got, err := client.ZRemRangeByLex(ctx, "ZRemRangeByLex-1", "-", "+")
		xt.NoError(t, err)
		xt.Equal(t, 0, got)

		num, err := client.ZAddMap(ctx, "ZRemRangeByLex-1", map[string]float64{"m1": 1, "m2": 2, "m3": 3})
		xt.NoError(t, err)
		xt.Equal(t, 3, num)

		got, err = client.ZRemRangeByLex(ctx, "ZRemRangeByLex-1", "-", "+")
		xt.NoError(t, err)
		xt.Equal(t, 3, got)
	})

	t.Run("ZRemRangeByRank", func(t *testing.T) {
		got, err := client.ZRemRangeByRank(ctx, "ZRemRangeByRank-1", 0, 1)
		xt.NoError(t, err)
		xt.Equal(t, 0, got)

		num, err := client.ZAddMap(ctx, "ZRemRangeByRank-1", map[string]float64{"m1": 1, "m2": 2, "m3": 3})
		xt.NoError(t, err)
		xt.Equal(t, 3, num)

		got, err = client.ZRemRangeByRank(ctx, "ZRemRangeByRank-1", 0, 1)
		xt.NoError(t, err)
		xt.Equal(t, 2, got)
	})

	t.Run("ZRemRangeByScore", func(t *testing.T) {
		got, err := client.ZRemRangeByScore(ctx, "ZRemRangeByScore-1", "-inf", "+inf")
		xt.NoError(t, err)
		xt.Equal(t, 0, got)

		num, err := client.ZAddMap(ctx, "ZRemRangeByScore-1", map[string]float64{"m1": 1, "m2": 2, "m3": 3})
		xt.NoError(t, err)
		xt.Equal(t, 3, num)

		got, err = client.ZRemRangeByScore(ctx, "ZRemRangeByScore-1", "-inf", "+inf")
		xt.NoError(t, err)
		xt.Equal(t, 3, got)
	})

	t.Run("ZScore", func(t *testing.T) {
		got, err := client.ZScore(ctx, "ZScore-1", "m1")
		xt.ErrorIs(t, err, ErrNil)
		xt.Equal(t, 0, got)

		num, err := client.ZAddMap(ctx, "ZScore-1", map[string]float64{"m1": 1, "m2": 2, "m3": 3})
		xt.NoError(t, err)
		xt.Equal(t, 3, num)

		got, err = client.ZScore(ctx, "ZScore-1", "m1")
		xt.NoError(t, err)
		xt.Equal(t, 1, got)
	})

	t.Run("ZScan", func(t *testing.T) {
		next, got, err := client.ZScan(ctx, "ZScan-1", 0, "", 0)
		xt.NoError(t, err)
		xt.Empty(t, got)
		xt.Equal(t, 0, next)

		num, err := client.ZAddMap(ctx, "ZScan-1", map[string]float64{"m1": 1, "m2": 2, "m3": 3})
		xt.NoError(t, err)
		xt.Equal(t, 3, num)

		next, got, err = client.ZScan(ctx, "ZScan-1", 0, "", 2)
		xt.NoError(t, err)
		xt.GreaterOrEqual(t, len(got), 2)
		xt.GreaterOrEqual(t, next, 0)

		var total int
		err = client.ZScanWalk(ctx, "ZScan-1", 0, "", 2, func(cursor uint64, m []Z) error {
			total += len(m)
			return nil
		})
		xt.NoError(t, err)
		xt.Equal(t, len(got), total)
	})
}
