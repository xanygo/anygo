//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-12-26

package xredis

import (
	"context"
	"fmt"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/store/xredis/resp3"
)

// https://redis.io/docs/latest/commands/cf.add/

// CFAdd 将指定元素添加到 Redis 的 Cuckoo Filter 中。
//
// 该方法对应 RedisBloom 模块的 CF.ADD 命令。
// Cuckoo Filter 是一种概率型数据结构，支持高效的插入、查询和删除操作。
// 如果元素已存在，Cuckoo Filter 仍会尝试添加。
//
// 参数：
//   - key: Cuckoo Filter 的名称，如果不存在，会自动创建
//   - item: 要添加的元素
//
// 返回值：
//   - bool: 如果元素成功添加返回 true，如果添加失败（如 Filter 已满）返回 false
func (c *Client) CFAdd(ctx context.Context, key string, item string) (bool, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBoolean, "CF.ADD", key, item)
	resp := c.do(ctx, cmd)
	return resp3.ToBool(resp.result, resp.err)
}

// CFAddNX 将指定元素添加到 Redis 的 Cuckoo Filter 中，仅在元素不存在时才会执行添加。
//
// 该方法对应 RedisBloom 模块的 CF.ADDNX 命令。
// 与 CFAdd 不同，CFAddNX 会先检查元素是否存在， 如果已经存在则不会重复添加。
//
// 参数：
//   - ctx: 上下文，用于控制超时或取消操作
//   - key: Cuckoo Filter 的名称，如果不存在，会自动创建
//   - item: 要添加的元素
//
// 返回值：
//   - bool: 如果元素成功添加返回 true；
//     如果元素已存在或添加失败（如 Filter 已满）返回 false
func (c *Client) CFAddNX(ctx context.Context, key string, item string) (bool, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBoolean, "CF.ADDNX", key, item)
	resp := c.do(ctx, cmd)
	return resp3.ToBool(resp.result, resp.err)
}

// CFCount 返回指定元素在 Redis 的 Cuckoo Filter 中的出现次数。
//
// 该方法对应 RedisBloom 模块的 CF.COUNT 命令。
// Cuckoo Filter 支持统计元素的频次，返回值为该元素在 filter 中的计数。
//
// 参数：
//   - key: Cuckoo Filter 的名称
//   - item: 要查询的元素
//
// 返回值：
//   - int64: 元素在 filter 中的出现次数
//   - 如果 key 或者 item 不存在，返回 0,nil
func (c *Client) CFCount(ctx context.Context, key string, item string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "CF.COUNT", key, item)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// CFDel 从 Redis 的 Cuckoo Filter 中删除指定元素。
//
// 该方法对应 RedisBloom 模块的 CF.DEL 命令。
// Cuckoo Filter 支持删除元素，如果元素存在，删除成功并更新内部计数。
// 如果元素不存在或删除失败，返回 false。
//
// 参数：
//   - ctx: 上下文，用于控制超时或取消操作
//   - key: Cuckoo Filter 的名称
//   - item: 要删除的元素
//
// 返回值：
//   - bool: 如果元素成功删除返回 true，
//   - 若 key 不存在 返回 false,error("ERR not found")
//   - 若 key 存在，但是 item 不存在，返回 false,nil
func (c *Client) CFDel(ctx context.Context, key string, item string) (bool, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBoolean, "CF.DEL", key, item)
	resp := c.do(ctx, cmd)
	return resp3.ToBool(resp.result, resp.err)
}

// CFExists 检查指定元素是否存在于 Redis 的 Cuckoo Filter 中。
//
// 该方法对应 RedisBloom 模块的 CF.EXISTS 命令。
// 如果元素存在于 filter 中，返回 true；否则返回 false。
// 注意：Cuckoo Filter 是概率型数据结构，在极少数情况下可能产生误判。
//
// 参数：
//   - ctx: 上下文，用于控制超时或取消操作
//   - key: Cuckoo Filter 的名称
//   - item: 要检查的元素
//
// 返回值：
//   - bool: true 表示该项很可能已经被添加到过滤器中; false 表示 key 不存在，或者 item 不存在。
func (c *Client) CFExists(ctx context.Context, key string, item string) (bool, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBoolean, "CF.EXISTS", key, item)
	resp := c.do(ctx, cmd)
	return resp3.ToBool(resp.result, resp.err)
}

// CFInFo 获取指定 Cuckoo Filter 的统计信息。
//
// 该方法对应 RedisBloom 模块的 CF.INFO 命令。
// 返回 filter 的容量、元素数量、桶大小、最大迭代次数等信息。
//
// 参数：
//   - key: Cuckoo Filter 的名称
//
// 返回值：
//   - Cuckoo Filter 的统计信息
//   - 若 key 不存在，返回 error("ERR not found")
func (c *Client) CFInFo(ctx context.Context, key string) (CFInfo, error) {
	cmd := resp3.NewRequest(resp3.DataTypeMap, "CF.INFO", key)
	resp := c.do(ctx, cmd)
	mp, err := resp3.ToStringAnyMap(resp.result, resp.err)
	if err != nil {
		return CFInfo{}, err
	}
	info := CFInfo{}
	xmap.Range(mp, func(key string, val int64) bool {
		switch key {
		case "Size":
			info.Size = val
		case "Number of buckets":
			info.NumBuckets = val
		case "Number of filters":
			info.NumFilters = val
		case "Number of items inserted":
			info.NumInserted = val
		case "Number of items deleted":
			info.NumDeleted = val
		case "Bucket size":
			info.BucketSize = val
		case "Expansion rate":
			info.ExpansionRate = val
		case "Max iterations":
			info.MaxIterations = val
		}
		return true
	})
	return info, nil
}

// CFInfo Cuckoo Filter 的统计信息
type CFInfo struct {
	Size          int64 // Filter 的容量（总桶数）
	NumBuckets    int64 // 桶的数量
	NumFilters    int64 // Filter 数量（通常为 1）
	NumInserted   int64 // 已插入元素数量
	NumDeleted    int64 // 已删除元素数量
	BucketSize    int64 // 每个桶的大小（存储槽数）
	ExpansionRate int64 // 自动扩容因子
	MaxIterations int64 // 插入元素时最大重排次数
}

// CFInsert 向 Cuckoo Filter 批量插入元素。
//
// 该方法对应 RedisBloom 模块的 CF.INSERT 命令，可以一次插入多个元素。
// 根据参数，可以控制是否在 Filter 不存在时创建，以及 Filter 的初始容量。
// 返回值为每个元素插入的结果，顺序与 items 一致。
//
// 参数：
//   - key: Cuckoo Filter 的名称
//   - items: 要插入的元素列表
//
// 返回值：
//   - []int64: 每个元素插入结果列表
//   - true 表示元素成功插入
//   - false 表示元素未插入（如已存在或 Filter 满）
func (c *Client) CFInsert(ctx context.Context, key string, items ...string) ([]bool, error) {
	return c.CFInsertWithOption(ctx, key, nil, items...)
}

// CFInsertWithOption 向 Cuckoo Filter 批量插入元素。
//
// 该方法对应 RedisBloom 模块的 CF.INSERT 命令，可以一次插入多个元素。
// 根据参数，可以控制是否在 Filter 不存在时创建，以及 Filter 的初始容量。
// 返回值为每个元素插入的结果，顺序与 items 一致。
//
// 参数：
//   - key: Cuckoo Filter 的名称
//   - items: 要插入的元素列表
//
// 返回值：
//   - []int64: 每个元素插入结果列表
//   - true 表示元素成功插入
//   - false 表示元素未插入（如已存在或 Filter 满）
func (c *Client) CFInsertWithOption(ctx context.Context, key string, opt *CFInsertOption, items ...string) ([]bool, error) {
	return c.doCFInsert(ctx, "CF.INSERT", key, opt, items...)
}

type CFInsertOption struct {
	// Capacity 当 Filter 不存在且需要创建时，指定初始容量,默认填写 0
	Capacity int64

	// NoCreate 如果为 true，则在 Filter 不存在时不创建，默认传 false
	NoCreate bool
}

func (c *Client) doCFInsert(ctx context.Context, command string, key string, opt *CFInsertOption, items ...string) ([]bool, error) {
	if len(items) == 0 {
		return nil, errNoMembers
	}
	args := []any{command, key}
	if opt != nil {
		if opt.Capacity > 0 {
			args = append(args, "CAPACITY", opt.Capacity)
		}
		if opt.NoCreate {
			args = append(args, "NOCREATE")
		}
	}

	args = append(args, "ITEMS")
	args = xslice.Append(args, items...)
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	arr, err := resp.asResp3Array(len(items))
	if err != nil {
		return nil, err
	}
	result := make([]bool, 0, len(items))
	// 返回结果,可能是bool，也可能是 1-添加成功，-1 失败（Filter 已满，容量不足）
	for _, v := range arr {
		switch tv := v.(type) {
		case resp3.Boolean:
			result = append(result, tv.Bool())
		case resp3.Integer:
			num := tv.Int()
			result = append(result, num == 1)
		default:
			return nil, fmt.Errorf("unexpected type %T", tv)
		}
	}
	return result, nil
}

// CFInsertNX 向 Redis 的 Cuckoo Filter 批量插入元素，仅在元素不存在时才插入。
//
// 该方法对应 RedisBloom 模块的 CF.INSERT 命令，并自动加上 NX 选项。
// 返回值为每个元素插入结果，顺序与 items 一致。
// 如果 Filter 已满或元素已存在，返回 false。
// 如果指定 NOCREATE 且 Filter 不存在，则返回错误。
//
// 参数：
//   - key: Cuckoo Filter 的名称
//   - items: 要插入的元素列表
//
// 返回值：
//   - []bool: 每个元素插入结果列表
//   - true 表示元素成功插入
//   - false 表示元素未插入（已存在或 Filter 已满）
func (c *Client) CFInsertNX(ctx context.Context, key string, items ...string) ([]bool, error) {
	return c.CFInsertNXWithOption(ctx, key, nil, items...)
}

// CFInsertNXWithOption 向 Redis 的 Cuckoo Filter 批量插入元素，仅在元素不存在时才插入。
//
// 该方法对应 RedisBloom 模块的 CF.INSERT 命令，并自动加上 NX 选项。
// 返回值为每个元素插入结果，顺序与 items 一致。
// 如果 Filter 已满或元素已存在，返回 false。
// 如果指定 NOCREATE 且 Filter 不存在，则返回错误。
//
// 参数：
//   - key: Cuckoo Filter 的名称
//   - items: 要插入的元素列表
//
// 返回值：
//   - []bool: 每个元素插入结果列表
//   - true 表示元素成功插入
//   - false 表示元素未插入（已存在或 Filter 已满）
func (c *Client) CFInsertNXWithOption(ctx context.Context, key string, opt *CFInsertOption, items ...string) ([]bool, error) {
	return c.doCFInsert(ctx, "CF.INSERTNX", key, opt, items...)
}

// CFLoadChunk 将 Cuckoo Filter 的序列化数据块加载到 Redis 中。
//
// 该方法对应 RedisBloom 模块的 CF.LOADCHUNK 命令。
// 通常用于从备份或持久化数据恢复 Cuckoo Filter。
// iterator 参数用于标记序列化数据的分块顺序，每次调用都会加载一块数据。
// 当 iterator 为 0 时表示开始加载新的 Filter。
//
// 参数：
//   - key: 目标 Cuckoo Filter 的名称
//   - iterator: 数据块序列号，用于标识分块顺序,由 CFScanDump 返回的值
//   - data: 序列化后的数据块字节数组
//
// 返回值：
//   - error: 执行过程中可能产生的错误，如果加载成功返回 nil
func (c *Client) CFLoadChunk(ctx context.Context, key string, iterator uint64, data []byte) error {
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, "CF.LOADCHUNK", key, iterator, data)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

// CFMExists 检查多个元素是否存在于 Redis 的 Cuckoo Filter 中。
//
// 该方法对应 RedisBloom 模块的 CFMEXISTS 命令。
// 可以一次性检查多个元素，返回值为每个元素的存在结果，顺序与 items 一致。
// 注意：Cuckoo Filter 是概率型数据结构，可能存在极少数误判。
//
// 参数：
//   - key: Cuckoo Filter 的名称
//   - items: 要检查的元素列表
//
// 返回值：
//   - []bool: 每个元素的存在状态
//   - true 表示元素存在
//   - false 表示 key 或者 item 不存在
func (c *Client) CFMExists(ctx context.Context, key string, items ...string) ([]bool, error) {
	if len(items) == 0 {
		return nil, errNoMembers
	}
	args := []any{"CF.MEXISTS", key}
	args = xslice.Append(args, items...)
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToBoolSlice(resp.result, resp.err, len(items))
}

// CFReserve 创建一个新的 Redis Cuckoo Filter，并为其分配指定容量。
//
// 该方法对应 RedisBloom 模块的 CF.RESERVE 命令。
// 可以通过 CFReserveOption 指定可选参数，如桶大小、扩容因子和最大重排次数。
// 如果 key 已存在，则返回错误。
//
// 参数：
//   - key: 要创建的 Cuckoo Filter 名称
//   - capacity: Filter 的初始容量（期望插入的元素数量）
//   - opt: 可选参数，指定桶大小、扩容因子和最大重排次数等
//
// 返回值：
//   - error: 执行过程中可能产生的错误，如果创建成功返回 nil
func (c *Client) CFReserve(ctx context.Context, key string, capacity int64, opt *CFReserveOption) error {
	args := []any{"CF.RESERVE", key, capacity}
	if opt != nil {
		args = opt.appendArgs(args)
	}
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

// CFReserveOption CFReserve 的可选参数。
//
// 该结构体用于 CF.RESERVE 命令，可用于控制 Filter 的内部配置。
// 如果某个字段未设置（为 0），RedisServer 会使用默认值。
type CFReserveOption struct {
	BucketSize    int64 // 每个桶的大小（每个桶可存储的槽数），默认值通常为 2
	MaxIterations int64 // 插入元素时的最大重排次数，默认值通常为 20
	Expansion     int64 // 自动扩容因子，默认值通常为 1
}

func (opt *CFReserveOption) appendArgs(args []any) []any {
	if opt.BucketSize > 0 {
		args = append(args, "BUCKETSIZE", opt.BucketSize)
	}
	if opt.MaxIterations > 0 {
		args = append(args, "MAXITERATIONS", opt.MaxIterations)
	}
	if opt.Expansion > 0 {
		args = append(args, "EXPANSION", opt.Expansion)
	}
	return args
}

// CFScanDump 从 Cuckoo Filter 中获取序列化的分块数据，用于备份或迁移。
//
// 该方法对应 RedisBloom 模块的 CF.SCANDUMP 命令。
// 每次调用返回一个数据块及下一个迭代器，直到迭代器返回 0 表示所有数据块已获取完毕。
// 通常与 CFLoadChunk 配合使用，用于将 Filter 从一个实例迁移到另一个实例。
//
// 参数：
//   - ctx: 上下文，用于控制超时或取消操作
//   - key: 目标 Cuckoo Filter 的名称
//   - iterator: 当前迭代器位置，第一次调用时传入 0
//
// 返回值：
//   - uint64: 下一个迭代器位置，如果为 0 表示已获取完所有数据块
//   - []byte: 当前迭代器对应的序列化数据块
func (c *Client) CFScanDump(ctx context.Context, key string, iterator uint64) (uint64, []byte, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "CF.SCANDUMP", key, iterator)
	resp := c.do(ctx, cmd)
	arr, err := resp.asResp3Array(2)
	if err != nil {
		return 0, nil, err
	}
	nextIterator, err1 := resp3.ToUint64(arr[0], nil)
	payload, err2 := resp3.ToBytes(arr[1], err1)
	return nextIterator, payload, err2
}
