//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-12-24

package xredis

import (
	"context"
	"fmt"
	"slices"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/store/xredis/resp3"
)

// https://redis.io/docs/latest/commands/bf.add/
// 依赖：redis-server  redisbloom 模块

// BFAdd 向布隆过滤器中添加一个元素
//
//	如果指定的 key 不存在，则会创建一个新的布隆过滤器，使用默认的误差率、容量和扩展参数
//	如果成功向过滤器中添加了一个元素，则返回 true；如果存在该元素可能已经被添加到过滤器中的概率，则返回 false
func (c *Client) BFAdd(ctx context.Context, key string, item string) (bool, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBoolean, "BF.ADD", key, item)
	resp := c.do(ctx, cmd)
	return resp3.ToBool(resp.result, resp.err)
}

// BFMAdd 向布隆过滤器中添加多个个元素
//
// 如果指定的 key 不存在，则会创建一个新的布隆过滤器，使用默认的误差率、容量和扩展参数
func (c *Client) BFMAdd(ctx context.Context, key string, items ...string) ([]bool, error) {
	if len(items) == 0 {
		return nil, errNoValues
	}
	args := []any{"BF.MADD", key}
	args = xslice.Append(args, items...)
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToBoolSlice(resp.result, resp.err)
}

// BFCard 返回布隆过滤器的基数：即被添加到布隆过滤器中并被判定为唯一的元素数量（指那些在至少一个子过滤器中至少设置了一个位的元素）
//
// 若 key 不存在，会返回 0
func (c *Client) BFCard(ctx context.Context, key string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "BF.CARD", key)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// BFExists 判断给定元素是否已被添加到布隆过滤器中
//
// 返回值：true 表示该元素很可能已经被添加到过滤器中；false 表示要么键不存在，要么该元素尚未被添加到过滤器中
func (c *Client) BFExists(ctx context.Context, key string, item string) (bool, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBoolean, "BF.EXISTS", key, item)
	resp := c.do(ctx, cmd)
	return resp3.ToBool(resp.result, resp.err)
}

// BFMExists 判断一个或多个元素是否已被添加到布隆过滤器
//
// 该命令与 BF.EXISTS 类似，但可以一次检查多个元素。
func (c *Client) BFMExists(ctx context.Context, key string, items ...string) ([]bool, error) {
	if len(items) == 0 {
		return nil, errNoValues
	}
	args := []any{"BF.MEXISTS", key}
	args = xslice.Append(args, items...)
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToBoolSlice(resp.result, resp.err)
}

func (c *Client) BFInfo(ctx context.Context, key string) (BFInfo, error) {
	cmd := resp3.NewRequest(resp3.DataTypeMap, "BF.INFO", key)
	resp := c.do(ctx, cmd)
	data, err := resp3.ToStringAnyMap(resp.result, resp.err)

	if err != nil {
		return BFInfo{}, resp.err
	}
	info := BFInfo{
		Capacity:  -1,
		Size:      -1,
		Filters:   -1,
		Items:     -1,
		Expansion: -1,
	}
	xmap.Range[string, int64](data, func(key string, val int64) bool {
		switch key {
		case "Capacity":
			info.Capacity = val
		case "Size":
			info.Size = val
		case "Number of filters":
			info.Filters = val
		case "Number of items inserted":
			info.Items = val
		case "Expansion rate":
			info.Expansion = val
		}
		return true
	})
	return info, nil
}

type BFInfo struct {
	// Capacity 在需要扩展之前，该布隆过滤器能够存储的唯一元素数量（包括已添加的元素）
	Capacity int64

	// Size 内存大小：为该布隆过滤器分配的字节数
	Size int64

	// Filters 子过滤器的数量
	Filters int64

	// Items 已添加到该布隆过滤器中并被判定为唯一的元素数量（即在至少一个子过滤器中至少设置了一个位的元素）
	Items int64

	// Expansion 膨胀率
	Expansion int64
}

// BFInsert 如果指定的键不存在，则使用指定的错误率、容量和扩展参数创建一个新的布隆过滤器，然后将所有指定的元素添加到该布隆过滤器中
//
//	key：要向其添加元素的布隆过滤器的名称。如果该 key 不存在，将创建一个新的布隆过滤器。
//	返回值：如果元素成功添加到过滤器中，则返回 true；如果元素很可能已存在于过滤器中，则返回 false
func (c *Client) BFInsert(ctx context.Context, key string, items []string, opt *BFInsertOption) ([]bool, error) {
	if len(items) == 0 {
		return nil, errNoValues
	}
	args := []any{"BF.INSERT", key}
	if opt != nil {
		args = opt.argsAppend(args)
	}
	args = slices.Grow(args, len(items)+1)
	args = append(args, "ITEMS")
	args = xslice.Append(args, items...)

	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToBoolSlice(resp.result, resp.err)
}

type BFInsertOption struct {
	// Capacity 指定新建过滤器的期望容量。如果过滤器已存在，该参数会被忽略。如果过滤器自动创建且未指定此参数，则使用模块级默认容量
	Capacity int

	// Error 指定新建过滤器的误差率（错误率），仅在过滤器尚不存在时生效。如果过滤器自动创建且未指定误差率，则使用模块级默认误差率
	Error float64

	// Expansion 当容量达到上限时，会创建一个额外的子过滤器。新子过滤器的大小 = 上一个子过滤器大小 × expansion（必须为正整数）。
	//
	//  如果不确定需要存储多少元素，建议将 expansion 设置为 2 或更大，以减少子过滤器数量。
	//  如果希望减少内存消耗，可将 expansion 设置为 1。
	//  默认值为 2
	Expansion int

	// NoCreate 如果过滤器不存在，则不创建新的过滤器。如果过滤器尚不存在，将返回错误，而不是自动创建。
	// 这可用于希望严格区分“创建过滤器”和“添加元素”操作的场景。
	// 注意：如果同时指定了 CAPACITY 或 ERROR，则使用 NOCREATE 会报错。
	NoCreate bool

	// NonScaling 禁止过滤器在达到初始容量后创建额外的子过滤器。
	// 非可扩展过滤器所需内存略少于可扩展过滤器。当容量达到上限时，过滤器会返回错误
	NonScaling bool
}

func (bo *BFInsertOption) argsAppend(args []any) []any {
	if bo.Capacity > 0 {
		args = append(args, "CAPACITY", bo.Capacity)
	}
	if bo.Error > 0 {
		args = append(args, "ERROR", bo.Error)
	}
	if bo.Expansion > 0 {
		args = append(args, "EXPANSION", bo.Expansion)
	}
	if bo.NoCreate {
		args = append(args, "NOCREATE")
	}
	if bo.NonScaling {
		args = append(args, "NONSCALING")
	}
	return args
}

// BFLoadChunk 恢复布隆过滤器：将之前使用 BF.SCANDUMP 保存的布隆过滤器恢复
func (c *Client) BFLoadChunk(ctx context.Context, key string, iterator int64, data []byte) (bool, error) {
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, "BF.LOADCHUNK", key, iterator, data)
	resp := c.do(ctx, cmd)
	return resp3.ToOkBool(resp.result, resp.err)
}

// BFScanDump 开始布隆过滤器的增量导出
//
//	该命令适用于无法一次性使用 DUMP 和 RESTORE 模型处理的大型布隆过滤器。
//	第一次调用该命令时，iter 的值应为 0。
//	该命令会依次返回 (iter, data) 对，直到返回 (0, NULL) 表示导出完成。
func (c *Client) BFScanDump(ctx context.Context, key string, iterator int64) (next int64, data []byte, err error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "BF.SCANDUMP", key, iterator)
	resp := c.do(ctx, cmd)
	arr, err := resp3.ToAnySlice(resp.result, resp.err)
	if err != nil {
		return 0, nil, err
	}
	if len(arr) != 2 {
		return 0, nil, fmt.Errorf("invalid array size %d", len(arr))
	}
	if first, ok := arr[0].(int64); ok {
		next = first
	} else {
		return 0, nil, fmt.Errorf("invalid first element %T, expect int64", arr[0])
	}

	if last, ok := arr[1].(string); ok {
		data = []byte(last)
	} else {
		return 0, nil, fmt.Errorf("invalid last element %T, expect string", arr[1])
	}
	return next, data, nil
}

// BFReserve 创建一个空的布隆过滤器，初始包含一个子过滤器，其容量为指定的初始容量，并且误差率不超过 error_rate。
//
//	默认情况下，当容量达到上限时，过滤器会自动扩展，通过创建额外的子过滤器来增加容量。新子过滤器的大小为上一个子过滤器大小乘以 expansion。
//	尽管过滤器可以通过创建子过滤器来扩展，但建议在创建时预留预计所需的容量。
//	原因是维护和查询子过滤器需要额外的内存（每个子过滤器都使用额外的位和哈希函数）并消耗更多 CPU 时间，而如果一开始就创建合适容量的过滤器，则效率更高。
//
// 哈希函数与位数计算：
//
//	最优哈希函数数量：ceil(-ln(error_rate) / ln(2))
//	每个元素所需的位数（给定误差率和最优哈希函数数量）：-ln(error_rate) / ln(2)^2
//	所需总位数：capacity * -ln(error_rate) / ln(2)^2
//
// 示例：
//
//	1% 误差率需要 7 个哈希函数，每个元素 9.585 位
//	0.1% 误差率需要 10 个哈希函数，每个元素 14.378 位
//	0.01% 误差率需要 14 个哈希函数，每个元素 19.170 位
func (c *Client) BFReserve(ctx context.Context, key string, errorRate int64, capacity int64, opt *BFReserveOption) error {
	args := []any{"BF.RESERVE", key, errorRate, capacity}
	if opt != nil {
		args = opt.argsAppend(args)
	}
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

type BFReserveOption struct {
	// Expansion 当容量达到上限时，会创建一个额外的子过滤器。新子过滤器的大小 = 上一个子过滤器的大小 × expansion（必须为正整数）。
	//
	//  如果不确定需要存储多少元素，建议将 expansion 设置为 2 或更大，以减少子过滤器数量。
	//  如果希望减少内存消耗，可将 expansion 设置为 1。
	//  默认值为 2。
	Expansion int

	//  NonScaling 当达到初始容量时，禁止过滤器创建额外的子过滤器。非可扩展（Non-scaling）过滤器所需内存略少于可扩展过滤器。当容量达到上限时，过滤器会返回错误。
	NonScaling bool
}

func (opt *BFReserveOption) argsAppend(args []any) []any {
	if opt.Expansion > 0 {
		args = append(args, "EXPANSION", opt.Expansion)
	}
	if opt.NonScaling {
		args = append(args, "NONSCALING")
	}
	return args
}
