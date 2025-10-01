//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-01

package xredis

import (
	"context"
	"time"

	"github.com/xanygo/anygo/store/xredis/resp3"
)

func (c *Client) Get(ctx context.Context, key string) StringResponse {
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, "GET", key)
	return StringResponse{
		base: c.do(ctx, cmd),
	}
}

func (c *Client) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	var args []any
	if ttl > 0 {
		args = []any{"SET", key, value, "PX", ttl.Milliseconds()}
	} else {
		args = []any{"SET", key, value}
	}
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, args...)
	resp := c.do(ctx, cmd)
	if resp.err != nil {
		return resp.err
	}
	return nil
}

// SetNX 当给定 key 不存在时，才设置 key 的值。设置成功返回 true，nil
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
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "INCRBYFLOAT", key, f)
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
	args := make([]any, len(keys)+1)
	args[0] = "MGET"
	for i, key := range keys {
		args[i+1] = key
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToMapWithKeys(resp.result, resp.err, keys)
}

func (c *Client) MSet(ctx context.Context, kv map[string]string) error {
	args := make([]any, 1, 2*len(kv)+1)
	args[0] = "MSET"
	for k, v := range kv {
		args = append(args, k, v)
	}
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, args...)
	resp := c.do(ctx, cmd)
	if resp.err != nil {
		return resp.err
	}
	return nil
}

func (c *Client) MSetNX(ctx context.Context, kv map[string]string) (int, error) {
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

func (c *Client) GetDel(ctx context.Context, key string) StringResponse {
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, "GETDEL", key)
	return StringResponse{
		base: c.do(ctx, cmd),
	}
}

func (c *Client) GetEx(ctx context.Context, key string, ttl time.Duration) StringResponse {
	var args []any
	if ttl > 0 {
		args = []any{"GETEX", key, "PX", ttl.Milliseconds()}
	} else {
		args = []any{"GETEX", key}
	}
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, args...)
	return StringResponse{
		base: c.do(ctx, cmd),
	}
}

func (c *Client) GetRange(ctx context.Context, key string, start int, end int) StringResponse {
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, "GETRANGE", key, start, end)
	return StringResponse{
		base: c.do(ctx, cmd),
	}
}

func (c *Client) GetSet(ctx context.Context, key string, value string) StringResponse {
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, "GETSET", key, value)
	return StringResponse{
		base: c.do(ctx, cmd),
	}
}
