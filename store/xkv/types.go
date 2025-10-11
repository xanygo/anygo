//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-20

package xkv

import (
	"context"

	"github.com/xanygo/anygo/store/xkv/internal"
)

var ErrInvalidType = internal.ErrInvalidType

type String[V any] interface {
	// Set 设置字符串的值（类似 Redis 的 SET 命令）
	Set(ctx context.Context, value V) error

	// Get 获取字符串的值（类似 Redis 的 GET 命令）
	// 返回：值，是否存在，错误
	Get(ctx context.Context) (V, bool, error)

	// Incr 将字符串中的数字自增 1（类似 Redis 的 INCR 命令）
	Incr(ctx context.Context) (int64, error)

	// Decr 将字符串中的数字自减 1（类似 Redis 的 DECR 命令）
	Decr(ctx context.Context) (int64, error)
}

type List[V any] interface {
	// LPush 在列表左侧插入元素（类似 Redis 的 LPUSH 命令）
	LPush(ctx context.Context, values ...V) (int, error)

	// RPush 在列表右侧插入元素（类似 Redis 的 RPUSH 命令）
	RPush(ctx context.Context, values ...V) (int, error)

	// LPop 移除并返回列表最左侧的元素（类似 Redis 的 LPOP 命令）
	LPop(ctx context.Context) (V, bool, error)

	// RPop 移除并返回列表最右侧的元素（类似 Redis 的 RPOP 命令）
	RPop(ctx context.Context) (V, bool, error)

	// LRem 从存储在键（key）的列表中删除等于元素（ element ）的前 count 个元素。count 参数以以下方式影响操作：
	// count > 0: 从头部到尾部移除等于 element 的元素。
	// count < 0: 从尾部到头部移除等于 element 的元素。
	// count = 0: 移除所有等于 element 的元素。
	// 例如，LREM list -2 "hello" 将从存储在 list 中的列表中删除 "hello" 的最后两个出现。
	// 请注意，不存在的键被视为空列表，因此当键不存在时，命令将始终返回0
	LRem(ctx context.Context, count int, element string) (int, error)

	// Range 不保证顺序的遍历
	Range(ctx context.Context, fn func(val V) bool) error

	LRange(ctx context.Context, fn func(val V) bool) error

	RRange(ctx context.Context, fn func(val V) bool) error
}

type Hash[V any] interface {
	// HSet 设置哈希表字段的值（类似 Redis 的 HSET 命令）
	HSet(ctx context.Context, field string, value V) error

	HMSet(ctx context.Context, data map[string]V) error

	// HGet 获取哈希表字段的值（类似 Redis 的 HGET 命令）
	// 返回：值，是否存在，错误
	HGet(ctx context.Context, field string) (V, bool, error)

	// HDel 删除哈希表中的某个字段（类似 Redis 的 HDEL 命令）
	HDel(ctx context.Context, fields ...string) error

	// HRange 遍历哈希表中的所有字段和
	HRange(ctx context.Context, fn func(field string, value V) bool) error

	// HGetAll 返回哈希表中的所有字段和值（类似 Redis 的 HGETALL 命令）
	HGetAll(ctx context.Context) (map[string]V, error)
}

// Set 无序、唯一元素集合
type Set[V any] interface {
	// SAdd 向集合中添加一个成员（类似 Redis 的 SADD 命令）
	// 返回值：新增个数，错误
	// 若 member 之前是存在的，则新增个数为 0
	SAdd(ctx context.Context, member ...V) (int, error)

	// SRem 从集合中移除一个成员（类似 Redis 的 SREM 命令）
	SRem(ctx context.Context, members ...V) error

	// SRange 遍历
	SRange(ctx context.Context, fn func(member V) bool) error

	// SMembers 返回集合中的所有成员（类似 Redis 的 SMEMBERS 命令）
	SMembers(ctx context.Context) ([]V, error)
}

// ZSet Sorted Set
type ZSet[V any] interface {
	// ZAdd 向有序集合中添加一个成员及其分数（类似 Redis 的 ZADD 命令）
	ZAdd(ctx context.Context, score float64, member V) error

	// ZScore 读取分数
	// 返回：值，是否存在，错误
	ZScore(ctx context.Context, member V) (float64, bool, error)

	// ZRange 按分数升序返回所有元素（类似 Redis 的 ZRANGE 命令）
	ZRange(ctx context.Context, fn func(member V, score float64) bool) error

	// ZRem 移除有序集合中的指定成员（类似 Redis 的 ZREM 命令）
	ZRem(ctx context.Context, members ...V) error
}

type Storage[V any] interface {
	String(key string) String[V]
	List(key string) List[V]
	Hash(key string) Hash[V]
	Set(key string) Set[V]
	ZSet(key string) ZSet[V]
	Delete(ctx context.Context, keys ...string) error
}

type StringStorage = Storage[string]
