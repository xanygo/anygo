//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-02

package xredis

import (
	"context"
	"time"

	"github.com/xanygo/anygo/store/xredis/resp3"
)

func (c *Client) HSet(ctx context.Context, key string, field, value string) (int, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "HSET", key, field, value)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

func (c *Client) HSetMap(ctx context.Context, key string, data map[string]string) (int, error) {
	if len(data) == 0 {
		return 0, errNoValues
	}
	args := make([]any, 2, 2*len(data)+2)
	args[0] = "HSET"
	args[1] = key
	for k, v := range data {
		args = append(args, k, v)
	}
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

// HSetEX 将给定哈希键的一个或多个字段设置为指定值，并可选择设置它们的过期时间或存活时间（TTL）
//
// redis server version >= 8.0.0
//
// opt: 可选值："FNX", "FXX"，“”:
//
//	 FNX: 仅当这些字段都不存在时才设置它们
//	 FXX: 仅当这些字段都已存在时才设置它们
//
//		ttl: 数据有效期, >0 时有效。等于 -1 时：保留字段原有的生存时间（TTL）
//
// 返回值：
// 1. 没有 field 写入，返回 false,nil
// 2. 全部 field 写入，返回 true,nil
// 3. 发生错误，返回  false,error
func (c *Client) HSetEX(ctx context.Context, key string, opt string, ttl time.Duration, data map[string]string) (bool, error) {
	if len(data) == 0 {
		return false, errNoValues
	}
	args := make([]any, 2, len(data)+2)
	args[0] = "HSETEX"
	args[1] = key
	switch opt {
	case "FNX", "FXX":
		args = append(args, opt)
	}
	if ttl > 0 {
		args = append(args, "PX", ttl.Milliseconds())
	} else if ttl == -1 {
		args = append(args, "KEEPTTL")
	}
	args = append(args, "FIELDS", len(data))
	for field, value := range data {
		args = append(args, field, value)
	}
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToIntBool(resp.result, resp.err, 1)
}

// HSetNX 将存储在指定键（key）中的哈希表里字段（field）的值设置为指定的值（value），仅当该字段尚不存在时才执行设置。
//
// 如果键（key）不存在，将会创建一个新的哈希表并保存该字段。
// 如果字段（field）已经存在，则此操作不会产生任何效果。
//
// 返回值：
//
//	1.当 field 作为新值被存储，返回 true,nil
//	2.当 field 已存在，则返回 false,nil
//	3.当发生错误，则返回 false,error
func (c *Client) HSetNX(ctx context.Context, key string, field string, value string) (bool, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "HSETNX", key, field, value)
	resp := c.do(ctx, cmd)
	return resp3.ToIntBool(resp.result, resp.err, 1)
}

// HStrLen 返回哈希表中指定字段的字符串值的长度。
//
// 参数 key 为哈希表键，field 为要查询的字段。
// 对应 Redis 的 HSTRLEN 命令。
func (c *Client) HStrLen(ctx context.Context, key string, field string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "HSTRLEN", key, field)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// HDel 删除哈希表中指定的一个或多个字段。
//
// 参数 key 为哈希表键，fields 为要删除的字段列表。
// 对应 Redis 的 HDEL 命令。
func (c *Client) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	args := make([]any, 2, len(fields)+2)
	args[0] = "HDEL"
	args[1] = key
	for _, field := range fields {
		args = append(args, field)
	}
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// HExists 检查哈希表中指定字段是否存在。
//
// 参数 key 为哈希表键，field 为要检查的字段。
// 对应 Redis 的 HEXISTS 命令。
func (c *Client) HExists(ctx context.Context, key string, field string) (bool, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "HEXISTS", key, field)
	resp := c.do(ctx, cmd)
	return resp3.ToIntBool(resp.result, resp.err, 1)
}

// HGet 返回存储在指定键（key）中的哈希表里，给定字段（field）所对应的值
//
// 返回值：
//  1. field 存在，返回 value，nil
//  2. field 不存在或 key 不存在，返回 “”，ErrNil
//  3. 其他错误，返回 "",error
func (c *Client) HGet(ctx context.Context, key string, field string) (string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, "HGET", key, field)
	resp := c.do(ctx, cmd)
	return resp3.ToString(resp.result, resp.err)
}

// HGetAll 返回存储在指定键（key）中的哈希表的所有字段（field）及其对应的值
//
// 若 key 不存在，会返回 nil，nil
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeMap, "HGETALL", key)
	resp := c.do(ctx, cmd)
	return resp3.ToStringMap(resp.result, resp.err)
}

// HGetDel 获取并删除指定哈希键（hash key）中的一个或多个字段（field）的值。当最后一个字段被删除时，该键（key）也会被删除。
func (c *Client) HGetDel(ctx context.Context, key string, fields ...string) (map[string]string, error) {
	args := make([]any, 4, len(fields)+4)
	args[0] = "HGETDEL"
	args[1] = key
	args[2] = "FIELDS"
	args[3] = len(fields)
	for _, field := range fields {
		args = append(args, field)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToStringMapWithKeys(resp.result, resp.err, fields)
}

// HGetEx 获取哈希表中指定字段的值，并可选择更新键的过期时间。
//
// 参数 key 为哈希表键，fields 为要获取的字段列表，必填。
// 参数 ttl 可指定过期时间：
//   - ttl > 0：将键的过期时间设置为指定毫秒数（PX）。
//   - ttl == -1：移除键的过期时间（PERSIST）。
//
// 对应 Redis 的 HGETEX 命令。
// 返回一个 map，包含字段及其对应的值；如果字段不存在，对应值为空字符串。
func (c *Client) HGetEx(ctx context.Context, key string, ttl time.Duration, fields ...string) (map[string]string, error) {
	if len(fields) == 0 {
		return nil, errNoFields
	}
	args := make([]any, 2, len(fields)+4)
	args[0] = "HGETEX"
	args[1] = key
	if ttl > 0 {
		args = append(args, "PX", ttl.Milliseconds())
	} else if ttl == -1 {
		args = append(args, "PERSIST")
	}
	args = append(args, "FIELDS", len(fields))
	for _, field := range fields {
		args = append(args, field)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToStringMapWithKeys(resp.result, resp.err, fields)
}

// HPersist 移除哈希键（hash key）中字段的现有过期时间，将这些字段从“可过期”（已设置过期时间）状态转换为“永久存在”（不再关联 TTL，永不过期）。
func (c *Client) HPersist(ctx context.Context, key string, fields ...string) (int, error) {
	args := make([]any, 4, len(fields)+4)
	args[0] = "HPERSIST"
	args[1] = key
	args[2] = "FIELDS"
	args[3] = len(fields)
	for _, field := range fields {
		args = append(args, field)
	}
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

// HIncrBy 将哈希表中指定字段的整数值增加指定增量。
//
// 参数 key 为哈希表键，field 为要增加的字段，increment 为增量值（可为负数）。
// 对应 Redis 的 HINCRBY 命令。
func (c *Client) HIncrBy(ctx context.Context, key string, field string, increment int) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "HINCRBY", key, field, increment)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// HIncrFloat 将哈希表中指定字段的浮点数值增加指定增量。
//
// 参数 key 为哈希表键，field 为要增加的字段，increment 为浮点数增量（可为负数）。
// 对应 Redis 的 HINCRBYFLOAT 命令。
func (c *Client) HIncrFloat(ctx context.Context, key string, field string, increment float64) (float64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, "HINCRBYFLOAT", key, field, increment)
	resp := c.do(ctx, cmd)
	return resp3.ToFloat64(resp.result, resp.err)
}

// HKeys 返回哈希表中所有字段的名称。
//
// 参数 key 为哈希表键。
// 对应 Redis 的 HKEYS 命令。
func (c *Client) HKeys(ctx context.Context, key string) ([]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "HKEYS", key)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}

// HLen 返回哈希表中字段的数量。
//
// 参数 key 为哈希表键。
// 对应 Redis 的 HLEN 命令。
// 如果哈希表为空或 key 不存在，返回 0。
func (c *Client) HLen(ctx context.Context, key string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "HLEN", key)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// HMGet 批量获取
//
//	若 key 不存在，会返回 nil,nil
//	若 field 不存在，则在返回的 map 中也不存在对应的 key
func (c *Client) HMGet(ctx context.Context, key string, fields ...string) (map[string]string, error) {
	if len(fields) == 0 {
		return nil, errNoFields
	}
	args := make([]any, 2, len(fields)+2)
	args[0] = "HMGET"
	args[1] = key
	for _, field := range fields {
		args = append(args, field)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToStringMapWithKeys(resp.result, resp.err, fields)
}

func (c *Client) HMSet(ctx context.Context, key string, data map[string]string) error {
	if len(data) == 0 {
		return errNoValues
	}
	args := make([]any, 2, len(data)+2)
	args[0] = "HMSET"
	args[1] = key
	for field, value := range data {
		args = append(args, field, value)
	}
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

// HPTTL 返回哈希表中一个或多个字段的剩余过期时间（TTL）。
//
// 参数 key 为哈希表键，fields 为要查询的字段列表。
// 对应 Redis 的 HPTTL 命令。
// 返回一个 time.Duration 切片，表示每个字段的剩余过期时间（毫秒）。
// 返回值说明：
//   - 正数：字段剩余过期时间（秒）
//   - -1：字段存在但未设置过期时间
//   - -2：字段不存在或 key 不存在
func (c *Client) HPTTL(ctx context.Context, key string, fields ...string) ([]time.Duration, error) {
	if len(fields) == 0 {
		return nil, errNoFields
	}
	args := make([]any, 4, len(fields)+2)
	args[0] = "HPTTL"
	args[1] = key
	args[2] = "FIELDS"
	args[3] = len(fields)
	for _, field := range fields {
		args = append(args, field, fields)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	list, err := resp3.ToInt64Slice(resp.result, resp.err)
	if err != nil {
		return nil, err
	}
	ret := make([]time.Duration, len(list))
	for i, num := range list {
		if num > 0 {
			ret[i] = time.Duration(num) * time.Millisecond
		} else {
			ret[i] = time.Duration(num)
		}
	}
	return ret, nil
}

// HTTL 返回哈希表中一个或多个字段的剩余过期时间（TTL），单位为秒。
//
// 参数 key 为哈希表键，fields 为要查询的字段列表。
// 对应 Redis 的 HTTL 命令。
// 返回一个 time.Duration 切片，表示每个字段的剩余过期时间（秒）。
// 返回值说明：
//   - 正数：字段剩余过期时间（秒）
//   - -1：字段存在但未设置过期时间
//   - -2：字段不存在或 key 不存在
func (c *Client) HTTL(ctx context.Context, key string, fields ...string) ([]time.Duration, error) {
	if len(fields) == 0 {
		return nil, errNoFields
	}
	args := make([]any, 4, len(fields)+2)
	args[0] = "HTTL"
	args[1] = key
	args[2] = "FIELDS"
	args[3] = len(fields)
	for _, field := range fields {
		args = append(args, field, fields)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	list, err := resp3.ToInt64Slice(resp.result, resp.err)
	if err != nil {
		return nil, err
	}
	ret := make([]time.Duration, len(list))
	for i, num := range list {
		if num > 0 {
			ret[i] = time.Duration(num) * time.Second
		} else {
			ret[i] = time.Duration(num)
		}
	}
	return ret, nil
}

// HVals 返回哈希表中所有字段的值。
//
// 参数 key 为哈希表键。
// 对应 Redis 的 HVALS 命令。
func (c *Client) HVals(ctx context.Context, key string) ([]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "HVALS", key)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}
