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

// TTL 返回一个已设置过期时间的键的剩余生存时间（精度：秒）
//
//	1.key 存在，但是无过期时间,返回 -1,nil
//	2.key 不存在,返回 0，ErrNil
//	3.有过期时间，返回 Duration,nil
//	4.出错，返回 0,error
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "TTL", key)
	resp := c.do(ctx, cmd)
	num, err := resp3.ToInt64(resp.result, resp.err)
	if err != nil {
		return 0, err
	}
	switch num {
	case -1:
		return -1, nil
	case -2:
		return 0, ErrNil
	default:
		return time.Duration(num) * time.Second, nil
	}
}

// PTTL 返回一个已设置过期时间的键的剩余生存时间（精度：毫秒）
//
//	1.key 存在，但是无过期时间,返回 -1,nil
//	2.key 不存在,返回 0，ErrNil
//	3.有过期时间，返回 Duration,nil
//	4.出错，返回 0,error
func (c *Client) PTTL(ctx context.Context, key string) (time.Duration, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "PTTL", key)
	resp := c.do(ctx, cmd)
	num, err := resp3.ToInt64(resp.result, resp.err)
	if err != nil {
		return 0, err
	}
	switch num {
	case -1:
		return -1, nil
	case -2:
		return 0, ErrNil
	default:
		return time.Duration(num) * time.Millisecond, nil
	}
}

func (c *Client) doKeysIntResult(ctx context.Context, method string, keys ...string) (int, error) {
	if len(keys) == 0 {
		return 0, errNoKeys
	}
	args := make([]any, 1, len(keys)+1)
	args[0] = method
	for _, key := range keys {
		args = append(args, key)
	}
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

// Del 删除指定的键。若键不存在，则会被忽略。返回值为删除的个数
func (c *Client) Del(ctx context.Context, keys ...string) (int, error) {
	return c.doKeysIntResult(ctx, "DEL", keys...)
}

// EXISTS 返回键是否存在。
//
// 需要注意的是，如果在参数中多次提到同一个已存在的键，该键会被重复计数。
// 例如：如果 somekey 存在，执行 EXISTS somekey somekey 将返回 2
func (c *Client) EXISTS(ctx context.Context, keys ...string) (int, error) {
	return c.doKeysIntResult(ctx, "EXISTS", keys...)
}

// Touch 修改一个或多个键的最后访问时间。若键不存在，则会被忽略。
func (c *Client) Touch(ctx context.Context, keys ...string) (int, error) {
	return c.doKeysIntResult(ctx, "TOUCH", keys...)
}

// Keys 返回所有与给定模式匹配的键
//
// 虽然该操作的时间复杂度为 O(N)，但常数时间开销相对较低。
// 例如，在一台入门级笔记本电脑上运行的 Redis，可以在约 40 毫秒内扫描包含 100 万个键的数据库。
//
// 支持的类似 glob 的模式：
//
//	h?llo     匹配 hello、hallo 和 hxllo
//	h*llo     匹配 hllo 和 heeeello
//	h[ae]llo  匹配 hello 和 hallo，但不匹配 hillo
//	h[^e]llo  匹配 hallo、hbllo 等，但不匹配 hello
//	h[a-b]llo 匹配 hallo 和 hbllo
func (c *Client) Keys(ctx context.Context, pattern string) ([]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "KEYS", pattern)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err)
}

// Move 将键从当前选定的数据库移动到指定的目标数据库。
//
// 如果目标数据库中已存在该键，或者源数据库中不存在该键，则不会执行任何操作。
// 由于这一特性，可以将 MOVE 用作一种简单的锁机制。
func (c *Client) Move(ctx context.Context, key string, db int) (bool, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "MOVE", key, db)
	resp := c.do(ctx, cmd)
	return resp3.ToIntBool(resp.result, resp.err, 1)
}

// Type 返回存储在键中的值的类型的字符串表示。
//
// 可能返回的类型包括：string、list、set、zset、hash、stream 和 vectorset
//
// 若 key 不存在，会返回 "none", ErrNil
func (c *Client) Type(ctx context.Context, key string) (string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, "TYPE", key)
	resp := c.do(ctx, cmd)
	tp, err := resp3.ToString(resp.result, resp.err)
	if err != nil {
		return "", err
	}
	if tp == "none" {
		return "none", ErrNil
	}
	return tp, nil
}

// Expire 为键设置超时。超时到期后，键将被自动删除 (时间精度：秒)
func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return c.ExpireOpt(ctx, key, ttl, "")
}

// ExpireOpt 为键设置超时 (时间精度：秒)
//
//	EXPIRE key seconds [NX | XX | GT | LT]
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
	return resp3.ToIntBool(resp.result, resp.err, 1)
}

// ExpireAt 的效果和语义与 EXPIRE 相同.(时间精度：秒)
//
// 如果指定的时间早于当前时间，键将立即被删除。
func (c *Client) ExpireAt(ctx context.Context, key string, at time.Time) (bool, error) {
	return c.ExpireAtOpt(ctx, key, at, "")
}

// ExpireAtOpt 给 key 设置过期截止时间 (时间精度：秒)
//
//	NX —— 仅在键没有过期时间时才设置过期时间
//	XX —— 仅在键已有过期时间时才设置过期时间
//	GT —— 仅在新过期时间大于当前过期时间时才设置过期时间
//	LT —— 仅在新过期时间小于当前过期时间时才设置过期时间
func (c *Client) ExpireAtOpt(ctx context.Context, key string, at time.Time, opt string) (bool, error) {
	var args []any
	if opt == "" {
		args = []any{"EXPIREAT", key, at.Unix()}
	} else {
		args = []any{"EXPIREAT", key, at.Unix(), opt}
	}
	return c.expire(ctx, args)
}

// ExpireTime 返回指定键的过期时间 (时间精度：秒)
//
// 1. 当 key 不存在时，返回 time.Time{},ErrNil
// 2. 当 key 存在，当无过期时间时，返回 time.Time{}(空的),nil
// 3. 当 key 存在，并且有过期时间，返回  time,nil
// 4. 当发生错误时，返回 time.Time{},error
func (c *Client) ExpireTime(ctx context.Context, key string) (time.Time, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "EXPIRETIME", key)
	resp := c.do(ctx, cmd)
	num, err := resp3.ToInt64(resp.result, resp.err)
	if err != nil {
		return time.Time{}, err
	}
	switch num {
	case -1:
		// -1 if the key exists but has no associated expiration time.
		return time.Time{}, nil
	case -2:
		// -2 if the key does not exist.
		return time.Time{}, ErrNil
	default:
		return time.Unix(num, 0), nil
	}
}

// PExpireTime 返回指定键的过期时间 (时间精度：毫秒)
//
// 1. 当 key 不存在时，返回 time.Time{},ErrNil
// 2. 当 key 存在，当无过期时间时，返回 time.Time{}(空的),nil
// 3. 当 key 存在，并且有过期时间，返回  time,nil
// 4. 当发生错误时，返回 time.Time{},error
func (c *Client) PExpireTime(ctx context.Context, key string) (time.Time, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "PEXPIRETIME", key)
	resp := c.do(ctx, cmd)
	num, err := resp3.ToInt64(resp.result, resp.err)
	if err != nil {
		return time.Time{}, err
	}
	switch num {
	case -1:
		// -1 if the key exists but has no associated expiration time.
		return time.Time{}, nil
	case -2:
		// -2 if the key does not exist.
		return time.Time{}, ErrNil
	default:
		return time.UnixMilli(num), nil
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

// Rename 将键重命名为 newkey。
//
// 如果原键不存在，则返回错误(ERR no such key)。
//
// 如果 newkey 已存在，它将被覆盖。在这种情况下，RENAME 会隐式执行一次 DEL 操作，
// 因此如果被删除的键包含非常大的值，可能会导致较高的延迟，尽管 RENAME 本身通常是一个常数时间操作。
func (c *Client) Rename(ctx context.Context, key string, newKey string) error {
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, "RENAME", key, newKey)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

// RenameNX 当 newkey 不存在时，将 key 重命名为 newkey。
//
// 如果原键不存在，则返回错误(ERR no such key)。
//
// 在集群模式下，key 和 newkey 必须位于同一个哈希槽中。
// 这意味着在实际使用中，只有具有相同哈希标签（hash tag）的键才能在集群中可靠地进行重命名。
//
//  1. 当 newkey 不存在，rename 成功时，返回 true，nil
//  2. 当 newkey 已存在时，返回 false,nil
//  3. 当 key 不存在时，或者其他异常时 ，返回 false,error
func (c *Client) RenameNX(ctx context.Context, key string, newKey string) (bool, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "RENAMENX", key, newKey)
	resp := c.do(ctx, cmd)
	return resp3.ToIntBool(resp.result, resp.err, 1)
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
