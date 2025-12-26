//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-12-26

package xredis

import (
	"context"
	"fmt"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/store/xredis/resp3"
)

// https://redis.io/docs/latest/commands/topk.add/

// TopKAdd 向指定 key 的 TopK 计数器添加一个元素。
//
// 参数说明：
//   - key: Redis 中的 TopK key 名称。
//   - item: 要添加的元素。
//
// 返回值：
//   - *string: 被淘汰的元素名称,如果未被淘汰则为 nil
//   - 若 key 不存在，会返回 nil,error("TopK: key does not exist")
func (c *Client) TopKAdd(ctx context.Context, key string, item string) (*string, error) {
	result, err := c.TopKAddN(ctx, key, item)
	if err != nil {
		return nil, err
	}
	return result[0], nil
}

// TopKAddN 向指定 key 的 TopK 计数器添加一个或多个元素。
//
// 参数说明：
//   - key: Redis 中的 TopK key 名称。
//   - items: 要添加的元素列表，可传入一个或多个元素。
//
// 返回值：
//   - []*string: 被淘汰的元素名称，顺序与输入 items 对应。
//     如果某个元素未被淘汰，对应位置为 nil。
func (c *Client) TopKAddN(ctx context.Context, key string, items ...string) ([]*string, error) {
	if len(items) == 0 {
		return nil, errNoValues
	}
	args := []any{"TOPK.ADD", key}
	args = xslice.Append(args, items...)
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToPtrStringSlice(resp.result, resp.err, len(items))
}

// TopKCount 查询指定 key 的单个元素在 TopK 计数器中的计数。
//
// 参数说明：
//   - key: Redis 中的 TopK key 名称。
//   - item: 要查询计数的元素名称。
//
// 返回值：
//   - int64: 元素在 TopK 中的计数，如果元素不存在，则返回 0
func (c *Client) TopKCount(ctx context.Context, key string, item string) (int64, error) {
	list, err := c.TopKCountN(ctx, key, item)
	if err != nil {
		return 0, err
	}
	return list[0], nil
}

// TopKCountN 批量查询指定 key 中多个元素的计数。
//
// 参数说明：
//   - key: Redis 中的 TopK key 名称。
//   - items: 要查询计数的元素列表，可传入一个或多个元素。
//
// 返回值：
//   - []int64: 与输入元素顺序一一对应，返回每个元素在 TopK 中的计数。
//     如果元素不存在，则返回 0。
func (c *Client) TopKCountN(ctx context.Context, key string, items ...string) ([]int64, error) {
	if len(items) == 0 {
		return nil, errNoValues
	}
	args := []any{"TOPK.COUNT", key}
	args = xslice.Append(args, items...)
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64Slice(resp.result, resp.err, len(items))
}

// TopKIncrBy 对指定 key 的单个 TopK 元素递增计数。
//
// 参数说明：
//   - key: Redis 中的 TopK key 名称。
//   - item: 要递增的元素名称。
//   - increment: 元素递增的数量。
//
// 返回值：
//   - *string: 如果递增操作导致某个元素被 TopK 淘汰，则返回被淘汰元素的名称；
//     如果没有元素被淘汰，返回 nil。
func (c *Client) TopKIncrBy(ctx context.Context, key string, item string, increment int64) (*string, error) {
	args := []any{"TOPK.INCRBY", key, item, increment}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	arr, err := resp3.ToPtrStringSlice(resp.result, resp.err, 1)
	if err != nil {
		return nil, err
	}
	return arr[0], nil
}

// TopKIncrByN 对指定 key 的 TopK 元素批量递增计数。
//
// 参数说明：
//   - key: Redis 中的 TopK key 名称。
//   - items: 要递增的元素列表，每个元素包含 Item 和 Increment 值，。
//
// 返回值：
//   - []*string: 返回每个递增操作可能被 TopK 丢弃的元素名称，顺序与输入 items 对应。
//     如果某个元素未被丢弃，对应位置为 nil。
func (c *Client) TopKIncrByN(ctx context.Context, key string, items ...TopKItemIncr) ([]*string, error) {
	if len(items) == 0 {
		return nil, errNoValues
	}
	args := []any{"TOPK.INCRBY", key}
	for _, item := range items {
		args = append(args, item.Item, item.Incr)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToPtrStringSlice(resp.result, resp.err, len(items))
}

// TopKItemIncr TopKIncrByN　方法使用
type TopKItemIncr struct {
	Item string
	Incr int64
}

// TopKInfo 返回指定 key 的 TopK 计数器信息。
//
// 参数说明：
//   - ctx: 上下文，用于控制超时或取消操作。
//   - key: Redis 中的 TopK key 名称。
//
// 返回值：
//   - TopKInfo: 返回 TopK 计数器的详细信息
func (c *Client) TopKInfo(ctx context.Context, key string) (TopKInfo, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "TOPK.INFO", key)
	resp := c.do(ctx, cmd)
	arr, err := resp.asResp3Array(0)
	if err != nil {
		return TopKInfo{}, err
	}
	if len(arr)%2 != 0 {
		return TopKInfo{}, fmt.Errorf("invalid TOPK.INFO reply: expected even number of elements, got %d", len(arr))
	}
	info := TopKInfo{}

	// > TOPK.INFO topk
	// 1# k => (integer) 50
	// 2# width => (integer) 2000
	// 3# depth => (integer) 7
	// 4# decay => (double) 0.925
	for i := 0; i < len(arr); i += 2 {
		name, err1 := resp3.ToString(arr[i], nil)
		if err1 != nil {
			return TopKInfo{}, err1
		}
		var err2 error
		switch name {
		case "k":
			info.K, err2 = resp3.ToInt64(arr[i+1], nil)
		case "width":
			info.Width, err2 = resp3.ToInt64(arr[i+1], nil)
		case "depth":
			info.Depth, err2 = resp3.ToInt64(arr[i+1], nil)
		case "decay":
			info.Decay, err2 = resp3.ToFloat64(arr[i+1], nil)
		}
		if err2 != nil {
			return info, err2
		}
	}

	return info, nil
}

type TopKInfo struct {
	K     int64   // top-k 的 k 值
	Width int64   // sketch 的宽度
	Depth int64   // sketch 的深度
	Decay float64 // 衰减因子
}

// TopKList 返回指定 key 的 TopK 元素列表。
//
// 参数说明：
//   - key: Redis 中的 TopK key 名称。
//
// 返回值：
//   - []string: 返回 TopK 元素的名称列表，顺序为从高到低频率。
func (c *Client) TopKList(ctx context.Context, key string) ([]string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "TOPK.LIST", key)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err, 0)
}

// TopKListWithCount 返回指定 key 的 TopK 元素及其计数。
//
// 参数说明：
//   - key: Redis 中的 TopK key 名称。
//
// 返回值：
//   - []TopKItemCount: 返回 TopK 元素及其对应的计数，顺序为从高到低频率。
func (c *Client) TopKListWithCount(ctx context.Context, key string) ([]TopKItemCount, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "TOPK.LIST", key, "WITHCOUNT")
	resp := c.do(ctx, cmd)
	arr, err := resp.asResp3Array(0)
	if err != nil {
		return nil, err
	}
	if len(arr)%2 != 0 {
		return nil, fmt.Errorf("invalid TOPK.LIST WITHCOUNT reply: expected even number of elements, got %d", len(arr))
	}
	result := make([]TopKItemCount, 0, len(arr)/2)
	for i := 0; i < len(arr); i += 2 {
		item, err1 := resp3.ToString(arr[i], nil)
		count, err2 := resp3.ToInt64(arr[i+1], err1)
		if err2 != nil {
			return nil, err2
		}
		result = append(result, TopKItemCount{Item: item, Count: count})
	}
	return result, nil
}

type TopKItemCount struct {
	Item  string
	Count int64 // 如果用户没传 WITHCOUNT，则 Count = 0
}

// TopKQuery 查询指定 key 的指定元素是否存在于 TopK 计数器中。
//
// 参数说明：
//   - key: Redis 中的 TopK key 名称。
//   - item: 要查询的元素列表。
//
// 返回值：
//
//	-bool: true 表示该元素在 TopK 中，false 表示该元素不在 TopK 中。
func (c *Client) TopKQuery(ctx context.Context, key string, item string) (bool, error) {
	result, err := c.TopKQueryN(ctx, key, item)
	if err != nil {
		return false, err
	}
	return result[0], nil
}

// TopKQueryN 查询指定 key 中多个元素是否存在于 TopK 计数器中。
//
// 参数说明：
//   - key: Redis 中的 TopK key 名称。
//   - items: 要查询的元素列表，可传入一个或多个元素。
//
// 返回值：
//   - []bool: 与输入元素顺序一一对应，true 表示该元素在 TopK 中，false 表示该元素不在 TopK 中。
func (c *Client) TopKQueryN(ctx context.Context, key string, items ...string) ([]bool, error) {
	if len(items) == 0 {
		return nil, errNoValues
	}
	args := []any{"TOPK.QUERY", key}
	args = xslice.Append(args, items...)
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToBoolSlice(resp.result, resp.err, len(items))
}

// TopKReserve 为指定的 key 创建或更新一个 TopK 计数器。
//
// 参数说明：
//   - key: Redis 中的 TopK key 名称。
//   - topK: 计数器保留的前 k 个最频繁元素。
//   - opt: 可选参数，用于设置 TopK 的额外配置，例如宽度、深度、衰减因子。
//     传入 nil 则使用默认配置。
//
// 返回值：
//   - error: 如果创建或更新失败，会返回错误；成功则返回 nil。
func (c *Client) TopKReserve(ctx context.Context, key string, topK int64, opt *TopKReserveOption) error {
	args := []any{"TOPK.RESERVE", key, topK}
	if opt != nil {
		args = opt.appendArgs(args)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

type TopKReserveOption struct {
	Width int64   // sketch 的宽度
	Depth int64   // sketch 的深度
	Decay float64 // 衰减因子
}

func (opt *TopKReserveOption) appendArgs(args []any) []any {
	return append(args, opt.Width, opt.Depth, opt.Decay)
}
