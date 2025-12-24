//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-01

package xredis

import (
	"context"
	"time"

	"github.com/xanygo/anygo/store/xredis/resp3"
)

func (c *Client) doKeyValuesIntResult(ctx context.Context, method string, key string, values ...string) (int64, error) {
	if len(values) == 0 {
		return 0, errNoValues
	}
	args := make([]any, 2, len(values)+2)
	args[0] = method
	args[1] = key
	for _, value := range values {
		args = append(args, value)
	}
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// LPush 将所有指定的值插入到存储在给定键的列表头部。 如果该键不存在，会在执行推入操作前创建一个空列表。
// 如果该键存在但其值不是列表类型，则返回错误。
//
// 返回值：在执行推入操作后，列表的长度
func (c *Client) LPush(ctx context.Context, key string, values ...string) (int64, error) {
	return c.doKeyValuesIntResult(ctx, "LPUSH", key, values...)
}

// LPushX 仅当给定的键已存在且其值为列表时，才将指定的值插入到列表头部。
// 与 LPUSH 不同的是，如果该键不存在，则不会执行任何操作。
//
// 返回值：在执行推入操作后，列表的长度
func (c *Client) LPushX(ctx context.Context, key string, values ...string) (int64, error) {
	return c.doKeyValuesIntResult(ctx, "LPUSHX", key, values...)
}

// RPush 将所有指定的值插入到存储在给定键的列表尾部。 如果该键不存在，会在执行推入操作前创建一个空列表。
// 如果该键存在但其值不是列表类型，则返回错误。
//
// 返回值：在执行推入操作后，列表的长度
func (c *Client) RPush(ctx context.Context, key string, values ...string) (int64, error) {
	return c.doKeyValuesIntResult(ctx, "RPUSH", key, values...)
}

// RPushX 仅当给定的键已存在且其值为列表时，才将指定的值插入到列表尾部。
// 与 RPUSH 不同的是，如果该键不存在，则不会执行任何操作。
//
// 返回值：在执行推入操作后，列表的长度
func (c *Client) RPushX(ctx context.Context, key string, values ...string) (int64, error) {
	return c.doKeyValuesIntResult(ctx, "RPUSHX", key, values...)
}

// LPop 移除并返回存储在给定键的列表的头部元素。
//
// 若 list 不存在，会返回 ErrNil
func (c *Client) LPop(ctx context.Context, key string) (string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, "LPOP", key)
	resp := c.do(ctx, cmd)
	return resp3.ToString(resp.result, resp.err)
}

// LPopN 移除并返回存储在给定键的列表的头部 最多 count 个元素。
//
// 若 list 不存在，会返回 ErrNil
//
//	若 list=[1, 2, 3, 4]，LPopN(list, 2) --> [1, 2]
func (c *Client) LPopN(ctx context.Context, key string, count int64) ([]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "LPOP", key, count)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}

// RPop 移除并返回存储在给定键的列表的尾部元素。
//
// 若 list 不存在，会返回 ErrNil
func (c *Client) RPop(ctx context.Context, key string) (string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, "RPOP", key)
	resp := c.do(ctx, cmd)
	return resp3.ToString(resp.result, resp.err)
}

// RPopN 移除并返回存储在给定键的列表的尾部最多 count 个元素。
//
// 若 list 不存在，会返回 ErrNil
//
//	若 list=[1, 2, 3, 4]，RPopN(list, 2) --> [4, 3]
func (c *Client) RPopN(ctx context.Context, key string, count int64) ([]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "RPOP", key, count)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}

func (c *Client) LLen(ctx context.Context, key string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "LLEN", key)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// LRem 从存储在键（key）的列表中删除等于元素（element）的前count个元素。count参数以以下方式影响操作：
// count > 0: 从头部到尾部移除等于element的元素。
// count < 0: 从尾部到头部移除等于element的元素。
// count = 0: 移除所有等于 element 的元素。
// 例如，LREM list -2 "hello" 将从存储在 list 中的列表中删除 "hello" 的最后两个出现。
// 请注意，不存在的键被视为空列表，因此当键不存在时，命令将始终返回0
func (c *Client) LRem(ctx context.Context, key string, count int64, element string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "LREM", key, count, element)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// LIndex https://redis.io/docs/latest/commands/lindex/
func (c *Client) LIndex(ctx context.Context, key string, index int64) (string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, "LINDEX", key, index)
	resp := c.do(ctx, cmd)
	return resp3.ToString(resp.result, resp.err)
}

func (c *Client) LInsertBefore(ctx context.Context, key string, pivot string, element string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "LINSERT", key, "BEFORE", pivot, element)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

func (c *Client) LInsertAfter(ctx context.Context, key string, pivot string, element string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "LINSERT", key, "AFTER", pivot, element)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

func (c *Client) LSet(ctx context.Context, key string, index int64, element string) error {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "LSET", key, index, element)
	resp := c.do(ctx, cmd)
	return resp.err
}

func (c *Client) LTrim(ctx context.Context, key string, start int64, stop int64) error {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "LTRIM", key, start, stop)
	resp := c.do(ctx, cmd)
	return resp.err
}

func (c *Client) bxPop(ctx context.Context, method string, timeout time.Duration, keys ...string) ([]string, error) {
	if len(keys) == 0 {
		return nil, errNoKeys
	}
	args := make([]any, 1, len(keys)+2)
	args[0] = method
	for _, key := range keys {
		args = append(args, key)
	}
	args = append(args, timeout.Seconds())
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}

// BRPop 是一个阻塞式的列表弹出原语。 它是 RPOP 的阻塞版本：当给定的任意列表都没有可弹出的元素时，连接会被阻塞。
// 当某个列表中有元素可供弹出时，会从第一个非空列表的尾部弹出一个元素，
// 列表的检查顺序按照命令中给定的键的顺序进行。
//
// 当 timeout 为 0 时，表示无限期阻塞。
func (c *Client) BRPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error) {
	return c.bxPop(ctx, "BRPOP", timeout, keys...)
}

// BLPop 是一个阻塞式的列表弹出原语。它是 LPOP 的阻塞版本：当给定的任意列表都没有可弹出的元素时，连接会被阻塞。
// 当某个列表中有元素可供弹出时，会从第一个非空列表的头部弹出一个元素，
// 列表的检查顺序按照命令中给定的键的顺序进行。
//
// 当 timeout 为 0 时，表示无限期阻塞。
func (c *Client) BLPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error) {
	return c.bxPop(ctx, "BLPOP", timeout, keys...)
}

// LRange 返回存储在指定键中的列表的指定元素。
//
//	参数 start 和 stop 表示偏移量（索引），它们是从零开始计数的：0 表示列表的第一个元素（表头），1 表示下一个元素，以此类推。
//	这些偏移量也可以是负数，表示从列表末尾开始的偏移。例如，-1 表示列表的最后一个元素，-2 表示倒数第二个元素，依此类推。
func (c *Client) LRange(ctx context.Context, key string, start int64, stop int64) ([]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "LRANGE", key, start, stop)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}
