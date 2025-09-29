//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-28

package xcache

import (
	"github.com/xanygo/anygo/internal/ztypes"
	"github.com/xanygo/anygo/xcodec"
)

type HasStats interface {
	Stats() Stats
}

type Stats struct {
	Get    uint64
	Set    uint64
	Delete uint64
	Hit    uint64
	Keys   int64 // 有多少个 key。-1: 未知,-2: 无有效的 Stats 信息
}

func (s Stats) String() string {
	str, _ := xcodec.EncodeToString(xcodec.JSON, s)
	return str
}

// HitRate 命中率
func (s Stats) HitRate() float64 {
	if s.Get == 0 {
		return 0
	}
	hit := float64(s.Hit) / float64(s.Get)
	return hit
}

const (
	statsKeysUnknown = -1 //  keys 数量未知
	statsKeysNoStats = -2 //  无有效的 Stats 信息
)

// GetStats 读取缓存对象的 统计信息
func GetStats(cache any) Stats {
	if hs, ok := cache.(HasStats); ok {
		return hs.Stats()
	}
	return Stats{
		Keys: statsKeysNoStats,
	}
}

type StatsRegistry ztypes.Registry[string, HasStats]

var statsRegistry StatsRegistry = ztypes.NewRegistry[string, HasStats]()

func Registry() StatsRegistry {
	return statsRegistry
}
