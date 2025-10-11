//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-02

package xredis

import (
	"context"
	"fmt"
	"time"

	"github.com/xanygo/anygo/store/xredis/resp3"
)

func (c *Client) HSet(ctx context.Context, key string, field, value string) (int, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "HSET", key, field, value)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

func (c *Client) HSetMap(ctx context.Context, key string, data map[string]string) (int, error) {
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
// opt: 可选值："FNX", "FXX"，“”。
// FNX: 仅当这些字段都不存在时才设置它们
// FXX: 仅当这些字段都已存在时才设置它们
//
//	ttl: 数据有效期, >0 时有效。等于 -1 时：保留字段原有的生存时间（TTL）
func (c *Client) HSetEX(ctx context.Context, key string, opt string, ttl time.Duration, data map[string]string) (int, error) {
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
	return resp3.ToInt(resp.result, resp.err)
}

func (c *Client) HSetNX(ctx context.Context, key string, field string, value string) (int, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "HSETNX", key, field, value)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

func (c *Client) HStrLen(ctx context.Context, key string, field string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "HSTRLEN", key, field)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

func (c *Client) HDel(ctx context.Context, key string, fields ...string) (int, error) {
	args := make([]any, 2, len(fields)+2)
	args[0] = "HDEL"
	args[1] = key
	for _, field := range fields {
		args = append(args, field)
	}
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

func (c *Client) HExists(ctx context.Context, key string, field string) (bool, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "HEXISTS", key, field)
	resp := c.do(ctx, cmd)
	num, err := resp3.ToInt(resp.result, resp.err)
	if err != nil {
		return false, err
	}
	switch num {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, fmt.Errorf("unexpect reply: %v", num)
	}
}

func (c *Client) HGet(ctx context.Context, key string, field string) (string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, "HGET", key, field)
	resp := c.do(ctx, cmd)
	return resp3.ToString(resp.result, resp.err)
}

func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeMap, "HGETALL", key)
	resp := c.do(ctx, cmd)
	return resp3.ToStringMap(resp.result, resp.err)
}

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

func (c *Client) HGetEx(ctx context.Context, key string, ttl time.Duration, fields ...string) (map[string]string, error) {
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

func (c *Client) HIncrBy(ctx context.Context, key string, field string, increment int) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "HINCRBY", key, field, increment)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

func (c *Client) HIncrFloat(ctx context.Context, key string, field string, increment float64) (float64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "HINCRBYFLOAT", key, field, increment)
	resp := c.do(ctx, cmd)
	return resp3.ToFloat64(resp.result, resp.err)
}

func (c *Client) HKeys(ctx context.Context, key string) ([]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "HKEYS", key)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}

func (c *Client) HLen(ctx context.Context, key string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "HLEN", key)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

func (c *Client) HMGet(ctx context.Context, key string, fields ...string) (map[string]string, error) {
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

func (c *Client) HPTTL(ctx context.Context, key string, fields ...string) ([]time.Duration, error) {
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

func (c *Client) HTTL(ctx context.Context, key string, fields ...string) ([]time.Duration, error) {
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

func (c *Client) HVals(ctx context.Context, key string) ([]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "HVALS", key)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}
