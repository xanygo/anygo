//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-01

package xredis

import (
	"context"
	"time"

	"github.com/xanygo/anygo/store/xredis/resp3"
)

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, "GET", key)
	resp := c.do(ctx, cmd)
	return resp3.ToString(resp.result, resp.err)
}

// Set 将键（key）设置为保存一个字符串值。如果该键已经持有某个值，则无论其类型是什么，都会被覆盖。
// 在 SET 操作成功后，之前与该键关联的任何生存时间（TTL）都会被清除。
func (c *Client) Set(ctx context.Context, key string, value any) error {
	return c.SetWithTTL(ctx, key, value, 0)
}

// SetWithTTL 将键（key）设置为保存一个字符串值。如果该键已经持有某个值，则无论其类型是什么，都会被覆盖。
// 在 SET 操作成功后，之前与该键关联的任何生存时间（TTL）都会被清除。
//
// 若不想设置 ttl，可以设置 ttl=0
func (c *Client) SetWithTTL(ctx context.Context, key string, value any, ttl time.Duration) error {
	var args []any
	if ttl > 0 {
		args = []any{"SET", key, value, "PX", ttl.Milliseconds()}
	} else {
		args = []any{"SET", key, value}
	}
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

// SetNX 当给定 key 不存在时，才设置 key 的值。
//
// 返回值有 3 种情况：
//
//	1.设置成功返回 true，nil
//	2.若 key 已存在 返回 false,nil
//	3.其他情况，返回 false,error
func (c *Client) SetNX(ctx context.Context, key string, value any, ttl time.Duration) (bool, error) {
	var args []any
	if ttl > 0 {
		args = []any{"SET", key, value, "NX", "PX", ttl.Milliseconds()}
	} else {
		args = []any{"SET", key, value, "NX"}
	}
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToOkBool(resp.result, resp.err)
}

// SetXX 当给定 key 存在时，才设置 key 的值。设置成功返回 true，nil
//
// 返回值有 3 种情况：
//
//	1.设置成功返回 true，nil
//	2.若 key 不存在 返回 false, nil
//	3.其他情况，返回 false, error
func (c *Client) SetXX(ctx context.Context, key string, value any, ttl time.Duration) (bool, error) {
	var args []any
	if ttl > 0 {
		args = []any{"SET", key, value, "XX", "PX", ttl.Milliseconds()}
	} else {
		args = []any{"SET", key, value, "XX"}
	}
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToOkBool(resp.result, resp.err)
}

func (c *Client) SetRange(ctx context.Context, key string, offset int, value any) (int, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "SETRANGE", key, offset, value)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

func (c *Client) StrLen(ctx context.Context, key string) (int, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "STRLEN", key)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "INCR", key)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

func (c *Client) IncrBy(ctx context.Context, key string, n int64) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "INCRBY", key, n)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

func (c *Client) IncrByFloat(ctx context.Context, key string, f float64) (float64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, "INCRBYFLOAT", key, f)
	resp := c.do(ctx, cmd)
	return resp3.ToFloat64(resp.result, resp.err)
}

func (c *Client) Decr(ctx context.Context, key string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "DECR", key)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

func (c *Client) DecrBy(ctx context.Context, key string, n int64) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "DECRBY", key, n)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// MGet 批量读取，若 key 不存在，在返回的 map 里也没有对应的 key
func (c *Client) MGet(ctx context.Context, keys ...string) (map[string]string, error) {
	if len(keys) == 0 {
		return nil, errNoKeys
	}
	args := make([]any, len(keys)+1)
	args[0] = "MGET"
	for i, key := range keys {
		args[i+1] = key
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToStringMapWithKeys(resp.result, resp.err, keys)
}

func (c *Client) MSet(ctx context.Context, kv map[string]string) error {
	if len(kv) == 0 {
		return errNoValues
	}
	args := make([]any, 1, 2*len(kv)+1)
	args[0] = "MSET"
	for k, v := range kv {
		args = append(args, k, v)
	}
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

func (c *Client) MSetNX(ctx context.Context, kv map[string]string) (int, error) {
	if len(kv) == 0 {
		return 0, errNoValues
	}
	args := make([]any, 1, 2*len(kv)+1)
	args[0] = "MSETNX"
	for k, v := range kv {
		args = append(args, k, v)
	}
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

func (c *Client) Append(ctx context.Context, key string, value string) (int, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "APPEND", key, value)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

// GetDel 获取指定键（key）的值并删除该键。
//
// 此命令类似于 GET，不同之处在于：在成功获取值后，会删除该键（仅当键的值类型为字符串时）。
func (c *Client) GetDel(ctx context.Context, key string) (string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, "GETDEL", key)
	resp := c.do(ctx, cmd)
	return resp3.ToString(resp.result, resp.err)
}

// GetEx 获取指定键（key）的值，同时可选择设置该键的过期时间。
//
// GETEX 类似于 GET，但它是一个 写操作命令，支持额外的选项来修改键的过期时间。
func (c *Client) GetEx(ctx context.Context, key string, ttl time.Duration) (string, error) {
	var args []any
	if ttl > 0 {
		args = []any{"GETEX", key, "PX", ttl.Milliseconds()}
	} else {
		args = []any{"GETEX", key}
	}
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToString(resp.result, resp.err)
}

// GetRange 返回存储在指定键中的字符串值的子串，由偏移量 start 和 end 决定（包含 end）。
//
// 可以使用负数偏移量，从字符串末尾开始计算。例如，-1 表示最后一个字符，-2 表示倒数第二个，以此类推。
// 如果请求的范围超出字符串长度，函数会自动将结果范围限制在字符串的实际长度内。
//
// https://redis.io/docs/latest/commands/getrange/
//
// 若 key 不存在，返回 “”，nil
// 若 end 超过字符串长度，会返回实际最大长度
func (c *Client) GetRange(ctx context.Context, key string, start int, end int) (string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, "GETRANGE", key, start, end)
	resp := c.do(ctx, cmd)
	return resp3.ToString(resp.result, resp.err)
}

// GetSet 原子地将指定键（key）设置为给定值（value），并返回该键原先存储的值。
//
// 如果键存在但不是字符串类型，则返回错误。如果设置成功，键之前的任何过期时间（TTL）都会保持。
func (c *Client) GetSet(ctx context.Context, key string, value string) (string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, "SET", key, value, "GET", "KEEPTTL")
	resp := c.do(ctx, cmd)
	return resp3.ToString(resp.result, resp.err)
}

func (c *Client) Echo(ctx context.Context, message string) (string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, "ECHO", message)
	resp := c.do(ctx, cmd)
	return resp3.ToString(resp.result, resp.err)
}
