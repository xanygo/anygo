//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-01

package xredis

import (
	"context"

	"github.com/xanygo/anygo/store/xredis/resp3"
)

func (c *Client) doKeyValuesIntResult(ctx context.Context, method string, key string, values ...string) (int, error) {
	args := make([]any, 2, len(values)+2)
	args[0] = method
	args[1] = key
	for _, value := range values {
		args = append(args, value)
	}
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

func (c *Client) LPush(ctx context.Context, key string, values ...string) (int, error) {
	return c.doKeyValuesIntResult(ctx, "LPUSH", key, values...)
}

func (c *Client) LPushX(ctx context.Context, key string, values ...string) (int, error) {
	return c.doKeyValuesIntResult(ctx, "LPUSHX", key, values...)
}

func (c *Client) RPush(ctx context.Context, key string, values ...string) (int, error) {
	return c.doKeyValuesIntResult(ctx, "RPUSH", key, values...)
}

func (c *Client) RPushX(ctx context.Context, key string, values ...string) (int, error) {
	return c.doKeyValuesIntResult(ctx, "RPUSHX", key, values...)
}

func (c *Client) LPop(ctx context.Context, key string) (string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, "LPOP", key)
	resp := c.do(ctx, cmd)
	return resp3.ToString(resp.result, resp.err)
}

func (c *Client) LPopN(ctx context.Context, key string, count int) ([]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "LPOP", key, count)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}

func (c *Client) RPop(ctx context.Context, key string) (string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, "RPOP", key)
	resp := c.do(ctx, cmd)
	return resp3.ToString(resp.result, resp.err)
}

func (c *Client) RPopN(ctx context.Context, key string, count int) ([]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "RPOP", key, count)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}

func (c *Client) LLen(ctx context.Context, key string) (int, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "LLEN ", key)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

func (c *Client) LRem(ctx context.Context, key string, count int, element string) (int, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "LREM ", key, count, element)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

// LIndex https://redis.io/docs/latest/commands/lindex/
func (c *Client) LIndex(ctx context.Context, key string, index int) (string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, "LINDEX ", key, index)
	resp := c.do(ctx, cmd)
	return resp3.ToString(resp.result, resp.err)
}

func (c *Client) LInsertBefore(ctx context.Context, key string, pivot string, element string) (int, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "LINSERT ", key, "BEFORE", pivot, element)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

func (c *Client) LInsertAfter(ctx context.Context, key string, pivot string, element string) (int, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "LINSERT ", key, "AFTER", pivot, element)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

func (c *Client) LSet(ctx context.Context, key string, index int, element string) error {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "LSET ", key, index, element)
	resp := c.do(ctx, cmd)
	return resp.err
}

func (c *Client) LTrim(ctx context.Context, key string, start int, stop int) error {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "LTRIM ", key, start, stop)
	resp := c.do(ctx, cmd)
	return resp.err
}

func (c *Client) bxPop(ctx context.Context, method string, timeout int, keys ...string) ([]string, error) {
	args := make([]any, 1, len(keys)+2)
	args[0] = method
	for _, key := range keys {
		args = append(args, key)
	}
	args = append(args, timeout)
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}

func (c *Client) BRPop(ctx context.Context, timeout int, keys ...string) ([]string, error) {
	return c.bxPop(ctx, "BRPOP", timeout, keys...)
}

func (c *Client) BLPop(ctx context.Context, timeout int, keys ...string) ([]string, error) {
	return c.bxPop(ctx, "BLPOP", timeout, keys...)
}
