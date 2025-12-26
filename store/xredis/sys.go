//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-12-26

package xredis

import (
	"context"
	"time"

	"github.com/xanygo/anygo/store/xredis/resp3"
)

// https://redis.io/docs/latest/commands/time/

// Time 返回 Redis 服务器当前的时间。
//
// 该方法对应 Redis 的 TIME 命令，用于获取服务器端的当前时间，而不是客户端本地时间。
//
// 返回值：
//   - time.Time: Redis 服务器当前时间（包含秒和微秒）
func (c *Client) Time(ctx context.Context) (time.Time, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "TIME")
	resp := c.do(ctx, cmd)
	arr, err := resp.asResp3Array(2)
	if err != nil {
		return time.Time{}, err
	}
	sec, err1 := resp3.ToInt64(arr[0], nil)
	micro, err2 := resp3.ToInt64(arr[1], err1)
	if err2 != nil {
		return time.Time{}, err2
	}
	return time.UnixMicro(sec*1e6 + micro), nil
}

// DBSize 返回当前 Redis 数据库中键的数量。
//
// 该方法对应 Redis 的 DBSIZE 命令，用于统计当前选中数据库（SELECT 之后）中
// 所有键的数量。该操作为 O(1)，返回的是键的总数，而不是内存占用大小。
//
// 返回值：
//   - int: 当前数据库中的键数量
func (c *Client) DBSize(ctx context.Context) (int, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "DBSIZE")
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

// LastSave 返回 Redis 最近一次成功执行持久化（RDB）保存的时间。
//
// 该方法对应 Redis 的 LASTSAVE 命令，用于获取服务器上一次成功生成 RDB 快照的时间点。
// 该时间由 Redis 服务器返回，通常用于监控持久化状态或判断数据是否已被持久化。
//
// 返回值：
//   - time.Time: 最近一次成功执行 RDB 保存的时间(精确到秒)
func (c *Client) LastSave(ctx context.Context) (time.Time, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "LASTSAVE")
	resp := c.do(ctx, cmd)
	sec, err := resp3.ToInt64(resp.result, resp.err)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(sec, 0), nil
}
