//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-20

package xkv

import "context"

type String interface {
	// Set 设置字符串的值（类似 Redis 的 SET 命令）
	Set(ctx context.Context, value string) error

	// Get 获取字符串的值（类似 Redis 的 GET 命令）
	Get(ctx context.Context) (string, error)

	// Incr 将字符串中的数字自增 1（类似 Redis 的 INCR 命令）
	Incr(ctx context.Context) (int64, error)

	// Decr 将字符串中的数字自减 1（类似 Redis 的 DECR 命令）
	Decr(ctx context.Context) (int64, error)
}

type List interface {
	// LPush 在列表左侧插入元素（类似 Redis 的 LPUSH 命令）
	LPush(ctx context.Context, val string) error

	// RPush 在列表右侧插入元素（类似 Redis 的 RPUSH 命令）
	RPush(ctx context.Context, val string) error

	// LPop 移除并返回列表最左侧的元素（类似 Redis 的 LPOP 命令）
	LPop(ctx context.Context) (string, bool, error)

	// RPop 移除并返回列表最右侧的元素（类似 Redis 的 RPOP 命令）
	RPop(ctx context.Context) (string, bool, error)

	// Range 不保证顺序的遍历
	Range(ctx context.Context, fn func(val string) bool) error

	LRange(ctx context.Context, fn func(val string) bool) error

	RRange(ctx context.Context, fn func(val string) bool) error
}

type Hash interface {
	// HSet 设置哈希表字段的值（类似 Redis 的 HSET 命令）
	HSet(ctx context.Context, field, value string) error

	// HGet 获取哈希表字段的值（类似 Redis 的 HGET 命令）
	HGet(ctx context.Context, field string) (string, bool, error)

	// HDel 删除哈希表中的某个字段（类似 Redis 的 HDEL 命令）
	HDel(ctx context.Context, field string) error

	// HRange 遍历哈希表中的所有字段和
	HRange(ctx context.Context, fn func(field, value string) bool) error

	// HGetAll 返回哈希表中的所有字段和值（类似 Redis 的 HGETALL 命令）
	HGetAll(ctx context.Context) (map[string]string, error)
}

// Set 无序、唯一元素集合
type Set interface {
	// SAdd 向集合中添加一个成员（类似 Redis 的 SADD 命令）
	SAdd(ctx context.Context, val string) error

	// SRem 从集合中移除一个成员（类似 Redis 的 SREM 命令）
	SRem(ctx context.Context, val string) error

	// SRange 遍历
	SRange(ctx context.Context, fn func(val string) bool) error

	// SMembers 返回集合中的所有成员（类似 Redis 的 SMEMBERS 命令）
	SMembers(ctx context.Context) ([]string, error)
}

// ZSet Sorted Set
type ZSet interface {
	// ZAdd 向有序集合中添加一个成员及其分数（类似 Redis 的 ZADD 命令）
	ZAdd(ctx context.Context, score float64, member string) error

	// ZScore 读取分数
	ZScore(ctx context.Context, member string) (float64, bool, error)

	// ZRange 按分数升序返回所有元素（类似 Redis 的 ZRANGE 命令）
	ZRange(ctx context.Context, fn func(member string, score float64) bool) error

	// ZRem 移除有序集合中的指定成员（类似 Redis 的 ZREM 命令）
	ZRem(ctx context.Context, member string) error
}

type Storage interface {
	String(key string) String
	List(key string) List
	Hash(key string) Hash
	Set(key string) Set
	ZSet(key string) ZSet
	Delete(ctx context.Context, key string) error
}
