//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-01

package xredis

import (
	"context"
	"fmt"
	"time"

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

func (c *Client) Touch(ctx context.Context, keys ...string) (int, error) {
	return c.doKeysIntResult(ctx, "TOUCH", keys...)
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

func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return c.ExpireOpt(ctx, key, ttl, "")
}

// ExpireOpt EXPIRE key seconds [NX | XX | GT | LT]
//
// NX —— 仅在键没有过期时间时才设置过期时间
// XX —— 仅在键已有过期时间时才设置过期时间
// GT —— 仅在新过期时间大于当前过期时间时才设置过期时间
// LT —— 仅在新过期时间小于当前过期时间时才设置过期时间
func (c *Client) ExpireOpt(ctx context.Context, key string, ttl time.Duration, opt string) (bool, error) {
	var args []any
	if opt == "" {
		args = []any{"EXPIRE", key, ttl.Seconds()}
	} else {
		args = []any{"EXPIRE", key, ttl.Seconds(), opt}
	}
	return c.expire(ctx, args)
}

func (c *Client) expire(ctx context.Context, args []any) (bool, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	num, err := resp3.ToInt(resp.result, resp.err)
	if err != nil {
		return false, err
	}
	switch num {
	case 1:
		return true, nil
	case 0:
		return false, ErrNil
	default:
		return false, fmt.Errorf("unknown reply value: %v", num)
	}
}

func (c *Client) ExpireAt(ctx context.Context, key string, at time.Time) (bool, error) {
	return c.ExpireAtOpt(ctx, key, at, "")
}

// ExpireAtOpt 给 key 设置过期截止时间
// NX —— 仅在键没有过期时间时才设置过期时间
// XX —— 仅在键已有过期时间时才设置过期时间
// GT —— 仅在新过期时间大于当前过期时间时才设置过期时间
// LT —— 仅在新过期时间小于当前过期时间时才设置过期时间
func (c *Client) ExpireAtOpt(ctx context.Context, key string, at time.Time, opt string) (bool, error) {
	var args []any
	if opt == "" {
		args = []any{"EXPIREAT", key, at.Unix()}
	} else {
		args = []any{"EXPIREAT", key, at.Unix(), opt}
	}
	return c.expire(ctx, args)
}

func (c *Client) ExpireTime(ctx context.Context, key string) (time.Time, bool, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "EXPIRETIME", key)
	resp := c.do(ctx, cmd)
	num, err := resp3.ToInt64(resp.result, resp.err)
	if err != nil {
		return time.Time{}, false, err
	}
	switch num {
	case -1:
		// -1 if the key exists but has no associated expiration time.
		return time.Time{}, false, nil
	case -2:
		// -2 if the key does not exist.
		return time.Time{}, false, ErrNil
	default:
		return time.Unix(num, 0), true, nil
	}
}

func (c *Client) PExpireTime(ctx context.Context, key string) (time.Time, bool, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "PEXPIRETIME", key)
	resp := c.do(ctx, cmd)
	num, err := resp3.ToInt64(resp.result, resp.err)
	if err != nil {
		return time.Time{}, false, err
	}
	switch num {
	case -1:
		// -1 if the key exists but has no associated expiration time.
		return time.Time{}, false, nil
	case -2:
		// -2 if the key does not exist.
		return time.Time{}, false, ErrNil
	default:
		return time.UnixMilli(num), true, nil
	}
}

// RandomKey 返回 DB 中随机一个 key
//
// 若数据库为空，会返回  ErrNil
func (c *Client) RandomKey(ctx context.Context) (string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, "RANDOMKEY")
	resp := c.do(ctx, cmd)
	return resp3.ToString(resp.result, resp.err)
}

func (c *Client) Rename(ctx context.Context, key string, newKey string) error {
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, "RENAME", key, newKey)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

func (c *Client) RenameNX(ctx context.Context, key string, newKey string) (bool, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "RENAMENX", key, newKey)
	resp := c.do(ctx, cmd)
	num, err := resp3.ToInt(resp.result, resp.err)
	if err != nil {
		return false, err
	}
	switch num {
	case 0:
		// 0 if newkey already exists.
		return false, nil
	case 1:
		// 1 if key was renamed to newkey.
		return true, nil
	default:
		return false, fmt.Errorf("unknown reply value: %v", num)
	}
}

func (c *Client) Scan(ctx context.Context, cursor uint64, match string, count int64, typ string) (*ScanResult, error) {
	sc := &ScanResult{
		cursor: cursor,
		match:  match,
		count:  count,
		typ:    typ,
		c:      c,
	}
	err := sc.Next(ctx)
	return sc, err
}

type ScanResult struct {
	cursor uint64
	match  string
	count  int64
	typ    string
	c      *Client

	resultKeys []string
}

func (sr *ScanResult) Next(ctx context.Context) error {
	args := make([]any, 2)
	args[0] = "SCAN"
	args[1] = sr.cursor
	if sr.match != "" {
		args = append(args, "MATCH", sr.match)
	}
	if sr.count > 0 {
		args = append(args, "COUNT", sr.count)
	}
	if sr.typ != "" {
		args = append(args, "TYPE", sr.typ)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := sr.c.do(ctx, cmd)
	if resp.err != nil {
		return resp.err
	}
	arr, ok := resp.result.(resp3.Array)
	if !ok || len(arr) != 2 {
		return fmt.Errorf("invalid result type: %T", resp.result)
	}
	first, err0 := resp3.ToInt64(arr[0], nil)
	if err0 != nil {
		return err0
	}

	elements, err1 := resp3.ToStringSlice(arr[1], nil)
	if err1 != nil {
		return err1
	}
	sr.cursor = uint64(first)
	sr.resultKeys = elements
	return nil
}

func (sr *ScanResult) Cursor() uint64 {
	return sr.cursor
}

func (sr *ScanResult) Keys() []string {
	return sr.resultKeys
}

func (sr *ScanResult) HasMore() bool {
	return sr.cursor > 0
}
