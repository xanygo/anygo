//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-01

package xredis

import (
	"context"

	"github.com/xanygo/anygo/store/xredis/resp3"
)

// TTL 返回一个已设置过期时间的键的剩余生存时间（单位：秒）
// -1: key 存在，但是无过期时间
// -2: key 不存在
func (c *Client) TTL(ctx context.Context, key string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "TTL", key)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// PTTL 返回一个已设置过期时间的键的剩余生存时间（单位：毫秒）
// -1: key 存在，但是无过期时间
// -2: key 不存在
func (c *Client) PTTL(ctx context.Context, key string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "PTTL", key)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

func (c *Client) doKeysIntResult(ctx context.Context, method string, keys ...string) (int, error) {
	args := make([]any, 1, len(keys)+1)
	args[0] = method
	for _, key := range keys {
		args = append(args, key)
	}
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

func (c *Client) Del(ctx context.Context, keys ...string) (int, error) {
	return c.doKeysIntResult(ctx, "DEL", keys...)
}

func (c *Client) EXISTS(ctx context.Context, keys ...string) (int, error) {
	return c.doKeysIntResult(ctx, "EXISTS", keys...)
}

func (c *Client) Keys(ctx context.Context, pattern string) ([]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "KEYS", pattern)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}

func (c *Client) Move(ctx context.Context, key string, db int) (bool, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "MOVE", key, db)
	resp := c.do(ctx, cmd)
	ret, err := resp3.ToInt(resp.result, resp.err)
	if err != nil {
		return false, err
	}
	return ret == 1, nil
}

func (c *Client) Type(ctx context.Context, key string) (string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, "TYPE", key)
	resp := c.do(ctx, cmd)
	return resp3.ToString(resp.result, resp.err)
}
