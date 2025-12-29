//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-12-27

package xredis

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/store/xredis/resp3"
)

// https://redis.io/docs/latest/commands/ts.add/

// TSAdd 向指定的时间序列中添加（或更新）一个样本点。
//
// 该方法对应 RedisTimeSeries 的 TS.ADD 命令，用于向时间序列写入
// 指定时间戳和数值的样本。如果时间序列不存在，将根据提供的可选参数
// 自动创建时间序列。
//
// 参数说明：
//   - key: 时间序列的键名。
//   - timestamp: 样本的时间戳（毫秒）。
//   - value: 样本的数值。
//   - opt: 可选参数，用于指定保留策略、编码方式、重复策略、忽略规则和标签等。
//
// 重复时间戳处理：
//   - 可以通过 DUPLICATE_POLICY 设置时间序列的默认重复策略。
//   - 可以通过 ON_DUPLICATE 为当前写入操作指定覆盖策略。
//   - 当两者同时存在时，ON_DUPLICATE 的优先级更高。
//
// 样本忽略行为：
//   - 当配置了 IGNORE 规则且样本被忽略时，不会写入新的数据点。
//   - 此时返回值为当前时间序列中最大的时间戳。
//
// 返回值：
//   - int64: 被插入（或更新）的样本时间戳。如果该样本被忽略，返回值将是该时间序列中当前最大的时间戳。
func (c *Client) TSAdd(ctx context.Context, key string, timestamp int64, value float64, opt *TSAddOption) (int64, error) {
	args := []any{"TS.ADD", key, timestamp, value}
	if opt != nil {
		args = opt.appendArgs(args)
	}
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// TSAddOption 定义 TS.ADD 命令的可选参数。
type TSAddOption struct {
	// Retention 指定时间序列的保留时间（毫秒）。超过该时间范围的数据点将被自动清理。
	// 只有当新创建时间序列的时候才有用。
	Retention int64

	// Encoding 指定时间序列的编码方式。只有当新创建时间序列的时候才有用。
	Encoding TSEncodingType

	// ChunkSize 指定每个时间序列数据块的大小（字节）。较大的块可提高压缩率，但会增加内存使用。
	// 只有当新创建时间序列的时候才有用。
	// 必须是 8 的倍数，并且位于 [48 .. 1048576]。redis-server 默认值为 4096
	ChunkSize int64

	// DuplicatePolicy 指定当写入的时间戳已存在时的处理策略。该策略会成为时间序列的默认重复策略。
	// 只有当新创建时间序列的时候才有用。
	DuplicatePolicy TSDuplicatePolicy

	// OnDuplicatePolicy 指定本次写入在发生时间戳重复时使用的覆盖策略。
	// 仅对当前 TS.ADD 调用生效，优先级高于 DuplicatePolicy。
	OnDuplicatePolicy TSDuplicatePolicy

	// Ignore 定义忽略样本的规则。当样本与已有数据的时间差或数值差超过指定阈值时，该样本会被忽略。
	Ignore *TSIgnoreRule

	// Labels 指定时间序列的标签集合。标签以键值对形式存储，用于过滤和聚合查询。
	// 只有当新创建时间序列的时候才有用。
	Labels map[string]string
}

func (opt *TSAddOption) appendArgs(args []any) []any {
	if opt.Retention > 0 {
		args = append(args, "RETENTION", opt.Retention)
	}

	if opt.Encoding != "" {
		args = append(args, "ENCODING", opt.Encoding)
	}

	if opt.ChunkSize > 0 {
		args = append(args, "CHUNK_SIZE", opt.ChunkSize)
	}

	if opt.DuplicatePolicy != "" {
		args = append(args, "DUPLICATE_POLICY", opt.DuplicatePolicy)
	}

	if opt.OnDuplicatePolicy != "" {
		args = append(args, "ON_DUPLICATE", opt.OnDuplicatePolicy)
	}

	if opt.Ignore != nil {
		args = append(args, "IGNORE", opt.Ignore.MaxTimeDiff, opt.Ignore.MaxValDiff)
	}

	if len(opt.Labels) > 0 {
		args = append(args, "LABELS")
		for k, v := range opt.Labels {
			args = append(args, k, v)
		}
	}
	return args
}

// TSEncodingType 表示时间序列的存储编码方式。
type TSEncodingType string

const (
	// TSEncodingCompressed 表示使用压缩编码方式存储时间序列数据。
	// 该方式可以显著减少内存占用，适合大多数场景。
	TSEncodingCompressed TSEncodingType = "COMPRESSED"

	// TSEncodingUncompressed 表示使用非压缩编码方式存储时间序列数据。
	// 该方式读写速度较快，但内存占用较高。
	TSEncodingUncompressed TSEncodingType = "UNCOMPRESSED"
)

// TSDuplicatePolicy 定义时间戳重复时的处理策略。
//
// 优先级说明：
//   - DUPLICATE_POLICY 用于设置时间序列的默认重复策略。
//   - ON_DUPLICATE 用于指定本次写入操作的重复策略，仅对当前命令生效。
//   - 当同时指定 ON_DUPLICATE 和 DUPLICATE_POLICY 时，
//     ON_DUPLICATE 的优先级更高，会覆盖 DUPLICATE_POLICY。
//
// 该类型用于 TS.ADD、TS.INCRBY、TS.DECRBY 等写入类命令。
type TSDuplicatePolicy string

const (
	// TSDuplicateBlock 表示当写入的时间戳已存在时直接返回错误。不会对已有样本进行修改。
	// 对应 RedisTimeSeries 的 BLOCK 策略。
	TSDuplicateBlock TSDuplicatePolicy = "BLOCK"

	// TSDuplicateFirst 表示保留第一次写入的值，忽略后续重复时间戳的写入。
	// 对应 RedisTimeSeries 的 FIRST 策略。
	TSDuplicateFirst TSDuplicatePolicy = "FIRST"

	// TSDuplicateLast 表示使用最新写入的值覆盖已有值。
	// 对应 RedisTimeSeries 的 LAST 策略。
	TSDuplicateLast TSDuplicatePolicy = "LAST"

	// TSDuplicateMin 表示在重复时间戳写入时，保留较小的值。
	// 对应 RedisTimeSeries 的 MIN 策略。
	TSDuplicateMin TSDuplicatePolicy = "MIN"

	// TSDuplicateMax 表示在重复时间戳写入时，保留较大的值。
	// 对应 RedisTimeSeries 的 MAX 策略。
	TSDuplicateMax TSDuplicatePolicy = "MAX"

	// TSDuplicateSum 表示在重复时间戳写入时，将新值与已有值相加后存储。
	// 对应 RedisTimeSeries 的 SUM 策略。
	TSDuplicateSum TSDuplicatePolicy = "SUM"
)

// TSIgnoreRule 定义样本忽略规则。
//
// 当新写入的样本与时间序列中已有数据点相比，时间差或数值差超过指定阈值时，该样本将被忽略，且不会写入时间序列。
//
// 该规则对应 RedisTimeSeries 中 TS.ADD / TS.CREATE 的 IGNORE 参数。
type TSIgnoreRule struct {
	// MaxTimeDiff 指定允许的最大时间差（毫秒）。
	// 当新样本的时间戳与最近样本的时间差超过该值时，样本将被忽略。
	MaxTimeDiff int64

	// MaxValDiff 指定允许的最大数值差。
	// 当新样本的值与最近样本的数值差超过该值时，样本将被忽略。
	MaxValDiff float64
}

// TSAlter 修改已存在时间序列的配置
//
// 该方法对应 RedisTimeSeries 的 TS.ALTER 命令，用于在不删除时间序列数据的情况下，更新其元数据和行为配置。
//
// TS.ALTER 仅影响后续写入的数据，不会对已存在的样本产生影响。
//
// 支持修改的配置包括：
//   - 保留时间（RETENTION）
//   - 重复时间戳处理策略（DUPLICATE_POLICY）
//   - 忽略规则（IGNORE）
//   - 标签（LABELS）
//
// 参数说明：
//   - key: 时间序列键名，必须已经存在。
//   - opt: 修改选项集合，至少需要指定一个字段。
//
// 返回值：
//   - error: 当时间序列不存在、参数非法或命令执行失败时返回错误。
//     当 key 不存在时，返回 “ERR TSDB: the key does not exist”
func (c *Client) TSAlter(ctx context.Context, key string, opt *TSAlterOption) error {
	if opt == nil {
		return errors.New("option is required")
	}
	args := []any{"TS.ALTER", key}
	args = opt.appendArgs(args)
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

// TSAlterOption 定义 TS.ALTER 命令的可选参数。
type TSAlterOption struct {
	// Retention 指定时间序列的保留时间（毫秒）。超过该时间范围的数据点将被自动清理。
	// 只有当新创建时间序列的时候才有用。
	Retention int64

	// ChunkSize 指定每个时间序列数据块的大小（字节）。较大的块可提高压缩率，但会增加内存使用。
	// 只有当新创建时间序列的时候才有用。
	// 必须是 8 的倍数，并且位于 [48 .. 1048576]。redis-server 默认值为 4096
	ChunkSize int64

	// DuplicatePolicy 指定当写入的时间戳已存在时的处理策略。该策略会成为时间序列的默认重复策略。
	// 只有当新创建时间序列的时候才有用。
	DuplicatePolicy TSDuplicatePolicy

	// Ignore 定义忽略样本的规则。当样本与已有数据的时间差或数值差超过指定阈值时，该样本会被忽略。
	Ignore *TSIgnoreRule

	// Labels 指定时间序列的标签集合。标签以键值对形式存储，用于过滤和聚合查询。
	// 只有当新创建时间序列的时候才有用。
	Labels map[string]string
}

func (opt *TSAlterOption) appendArgs(args []any) []any {
	if opt.Retention > 0 {
		args = append(args, "RETENTION", opt.Retention)
	}

	if opt.ChunkSize > 0 {
		args = append(args, "CHUNK_SIZE", opt.ChunkSize)
	}

	if opt.DuplicatePolicy != "" {
		args = append(args, "DUPLICATE_POLICY", opt.DuplicatePolicy)
	}

	if opt.Ignore != nil {
		args = append(args, "IGNORE", opt.Ignore.MaxTimeDiff, opt.Ignore.MaxValDiff)
	}

	if len(opt.Labels) > 0 {
		args = append(args, "LABELS")
		for k, v := range opt.Labels {
			args = append(args, k, v)
		}
	}
	return args
}

// TSCreate 创建一个新的时间序列。
//
// 该方法对应 RedisTimeSeries 的 TS.CREATE 命令，用于创建一个时间序列键，并定义其存储、编码和写入行为。
//
// 参数说明：
//   - key: 时间序列键名，若键已存在将返回错误。
//   - opt: 创建选项集合，可为 nil 表示使用默认配置。
//
// 返回值：
//   - error: 当键已存在、参数非法或命令执行失败时返回错误。
//   - 当 key 已存在时返回错误：ERR TSDB: key already exists
func (c *Client) TSCreate(ctx context.Context, key string, opt *TTSCreateOption) error {
	args := []any{"TS.CREATE", key}
	if opt != nil {
		args = opt.appendArgs(args)
	}
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

// TTSCreateOption 定义 TS.ADD 命令的可选参数。
type TTSCreateOption struct {
	// Retention 指定时间序列的保留时间（毫秒）。超过该时间范围的数据点将被自动清理。
	// 只有当新创建时间序列的时候才有用。
	Retention int64

	// Encoding 指定时间序列的编码方式。只有当新创建时间序列的时候才有用。
	Encoding TSEncodingType

	// ChunkSize 指定每个时间序列数据块的大小（字节）。较大的块可提高压缩率，但会增加内存使用。
	// 只有当新创建时间序列的时候才有用。
	// 必须是 8 的倍数，并且位于 [48 .. 1048576]。redis-server 默认值为 4096
	ChunkSize int64

	// DuplicatePolicy 指定当写入的时间戳已存在时的处理策略。该策略会成为时间序列的默认重复策略。
	// 只有当新创建时间序列的时候才有用。
	DuplicatePolicy TSDuplicatePolicy

	// Ignore 定义忽略样本的规则。当样本与已有数据的时间差或数值差超过指定阈值时，该样本会被忽略。
	Ignore *TSIgnoreRule

	// Labels 指定时间序列的标签集合。标签以键值对形式存储，用于过滤和聚合查询。
	// 只有当新创建时间序列的时候才有用。
	Labels map[string]string
}

func (opt *TTSCreateOption) appendArgs(args []any) []any {
	if opt.Retention > 0 {
		args = append(args, "RETENTION", opt.Retention)
	}

	if opt.Encoding != "" {
		args = append(args, "ENCODING", opt.Encoding)
	}

	if opt.ChunkSize > 0 {
		args = append(args, "CHUNK_SIZE", opt.ChunkSize)
	}

	if opt.DuplicatePolicy != "" {
		args = append(args, "DUPLICATE_POLICY", opt.DuplicatePolicy)
	}

	if opt.Ignore != nil {
		args = append(args, "IGNORE", opt.Ignore.MaxTimeDiff, opt.Ignore.MaxValDiff)
	}

	if len(opt.Labels) > 0 {
		args = append(args, "LABELS")
		for k, v := range opt.Labels {
			args = append(args, k, v)
		}
	}
	return args
}

// TSCreateRule 创建时间序列的压缩（聚合）规则。
//
// 该方法对应 RedisTimeSeries 的 TS.CREATERULE 命令，用于在源时间序列与目标时间序列之间建立聚合规则，
// 使写入源序列的数据按指定规则自动聚合并写入目标序列。
//
// 参数说明：
//   - sourceKey: 源时间序列键名，写入该序列的数据将触发聚合。
//   - destKey: 目标时间序列键名，聚合结果将写入该序列。
//   - aggregator: 聚合函数类型（如 AVG、SUM、MIN、MAX、COUNT 等）。
//   - bucketDuration: 聚合桶的时间跨度，单位为毫秒。
//   - alignTimestamp: 可选的对齐时间戳，用于确定聚合桶的起始对齐点；为 nil 时表示不指定对齐时间戳，由 Redis 使用默认对齐规则。
//
// 返回值：
//   - error: 当源或目标时间序列不存在、参数非法、
//     规则已存在或命令执行失败时返回错误。
//
// 注意事项：
//   - 一个源时间序列可以关联多个聚合规则。
//   - 目标时间序列必须已存在，且不能与源时间序列相同。
//   - alignTimestamp 用于对齐聚合桶的边界，其取值应为合法的
//     Unix 时间戳（毫秒）。
//     当 sourceKey 或者 destKey 不存在时，会返回错误：ERR TSDB: the key does not exist
func (c *Client) TSCreateRule(ctx context.Context, sourceKey, destKey string, aggregator string, bucketDuration int64, alignTimestamp *int64) error {
	args := []any{"TS.CREATERULE", sourceKey, destKey, "AGGREGATION", aggregator, bucketDuration}
	if alignTimestamp != nil {
		args = append(args, *alignTimestamp)
	}
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

// TSDecrBy 对指定时间序列的最新值执行减操作。
//
// 该方法对应 RedisTimeSeries 的 TS.DECRBY 命令，用于将时间序列当前最新样本的值减少指定的 subtrahend，并将结果作为新的样本写入。
// 如果时间序列不存在或者没有最新样本，可以根据 opt 创建新样本。
//
// 参数说明：
//   - key: 时间序列键名，必须已存在。
//   - subtrahend: 要从最新值中减去的数值。
//   - opt: 可选参数，用于控制写入行为（如时间戳、RETENTION、ON_DUPLICATE、IGNORE 等）。
//
// 返回值：
//   - int64: 实际写入的样本时间戳（毫秒级 Unix 时间戳）。
//   - error: 当时间序列不存在时会创建新的。
func (c *Client) TSDecrBy(ctx context.Context, key string, subtrahend float64, opt *TSDecrByOption) (int64, error) {
	args := []any{"TS.DECRBY", key, subtrahend}
	if opt != nil {
		args = opt.appendArgs(args)
	}
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// TSDecrByOption 定义 TS.DECRBY 命令的可选参数。
type TSDecrByOption struct {
	// TimeStamp  是 Unix 时间（毫秒级整数），用于指定样本的时间戳, 值 >0 时有效。
	//
	// Unix 时间是自 1970 年 1 月 1 日 00:00:00 UTC（Unix 纪元）以来经过的毫秒数，不包含闰秒调整。
	//
	// timestamp 必须等于或大于当前已存在的最大时间戳。
	// 当 timestamp 等于最大已存在时间戳时，会减少该最大时间戳对应样本的值。
	// 当 timestamp 大于最大已存在时间戳时，会创建一个新的样本，其时间戳设置为 timestamp，其值设置为“最大已存在时间戳对应样本的值减去 subtrahend”。
	//
	// 如果未指定时间戳（TimeStamp <= 0 ），则时间戳会被设置为服务器当前时钟的 Unix 时间。
	TimeStamp int64

	// Retention 指定时间序列的保留时间（毫秒）。超过该时间范围的数据点将被自动清理。
	// 只有当新创建时间序列的时候才有用。
	Retention int64

	// Encoding 指定时间序列的编码方式。只有当新创建时间序列的时候才有用。
	Encoding TSEncodingType

	// ChunkSize 指定每个时间序列数据块的大小（字节）。较大的块可提高压缩率，但会增加内存使用。
	// 只有当新创建时间序列的时候才有用。
	// 必须是 8 的倍数，并且位于 [48 .. 1048576]。redis-server 默认值为 4096
	ChunkSize int64

	// DuplicatePolicy 指定当写入的时间戳已存在时的处理策略。该策略会成为时间序列的默认重复策略。
	// 只有当新创建时间序列的时候才有用。
	DuplicatePolicy TSDuplicatePolicy

	// Ignore 定义忽略样本的规则。当样本与已有数据的时间差或数值差超过指定阈值时，该样本会被忽略。
	Ignore *TSIgnoreRule

	// Labels 指定时间序列的标签集合。标签以键值对形式存储，用于过滤和聚合查询。
	// 只有当新创建时间序列的时候才有用。
	Labels map[string]string
}

func (opt *TSDecrByOption) appendArgs(args []any) []any {
	if opt.TimeStamp == 0 {
		args = append(args, "TIMESTAMP", "*")
	} else if opt.TimeStamp > 0 {
		args = append(args, "TIMESTAMP", opt.TimeStamp)
	}
	if opt.TimeStamp > 0 {
		args = append(args, "TIMESTAMP", opt.TimeStamp)
	}

	if opt.Retention > 0 {
		args = append(args, "RETENTION", opt.Retention)
	}

	if opt.Encoding != "" {
		args = append(args, "ENCODING", opt.Encoding)
	}

	if opt.ChunkSize > 0 {
		args = append(args, "CHUNK_SIZE", opt.ChunkSize)
	}

	if opt.DuplicatePolicy != "" {
		args = append(args, "DUPLICATE_POLICY", opt.DuplicatePolicy)
	}

	if opt.Ignore != nil {
		args = append(args, "IGNORE", opt.Ignore.MaxTimeDiff, opt.Ignore.MaxValDiff)
	}

	if len(opt.Labels) > 0 {
		args = append(args, "LABELS")
		for k, v := range opt.Labels {
			args = append(args, k, v)
		}
	}
	return args
}

// TSDel 用于删除时间序列中指定时间范围内的样本。
//
// 该方法会删除 key 对应的时间序列中，时间戳位于
// [fromTimestamp, toTimestamp]（包含边界）范围内的所有样本。
//
// 参数说明：
//   - key: 时间序列的键名。
//   - fromTimestamp: 起始时间戳（毫秒级 Unix 时间戳）。
//   - toTimestamp: 结束时间戳（毫秒级 Unix 时间戳）。
//
// 返回值：
//   - int64: 实际被删除的样本数量。
//   - 若 key 不存在，会返回错误：ERR TSDB: the key does not exist
func (c *Client) TSDel(ctx context.Context, key string, fromTimestamp int64, toTimestamp int64) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "TS.DEL", key, fromTimestamp, toTimestamp)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// TSDelRule 用于删除时间序列之间已存在的压缩规则（compaction rule）。
//
// 该方法会删除从 sourceKey 指向 destKey 的时间序列规则，
// 规则删除后，sourceKey 的数据将不再自动聚合并写入 destKey。
//
// 参数说明：
//   - sourceKey: 源时间序列的键名。
//   - destKey: 目标时间序列的键名。
//
// 若 sourceKey 或者 destKey 不存在，会返回错误： ERR TSDB: the key does not exist
func (c *Client) TSDelRule(ctx context.Context, sourceKey string, destKey string) error {
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, "TS.DELETERULE", sourceKey, destKey)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

// TSGet 用于获取时间序列的最新样本。
//
// 该方法返回 key 对应时间序列中最新一条样本的数据，通常包括样本的时间戳和值。
//
// 参数说明：
//   - key: 时间序列的键名。
//
// 返回值：
//   - *TSItem: 最新样本项，包含时间戳和值等信息。
//   - 若 key 不存在，会返回错误： ERR TSDB: the key does not exist
func (c *Client) TSGet(ctx context.Context, key string) (*TSSample, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "TS.GET", key)
	resp := c.do(ctx, cmd)
	arr, err := resp.asResp3Array(0)
	if err != nil {
		return nil, err
	}
	if len(arr) == 0 {
		return nil, nil
	}
	if len(arr) != 2 {
		return nil, fmt.Errorf("invalid array length: %d", len(arr))
	}
	timestamp, err1 := resp3.ToInt64(arr[0], nil)
	value, err2 := resp3.ToFloat64(arr[1], err1)
	if err2 != nil {
		return nil, err2
	}
	return &TSSample{Timestamp: timestamp, Value: value}, nil
}

type TSItem struct {
	Key       string
	Timestamp int64
	Value     float64
}

// TSSample 表示时间序列中的单个采样点。
type TSSample struct {
	Timestamp int64
	Value     float64
}

// TSIncrBy 用于将时间序列的最新值按给定增量进行递增。
//
// 该方法会对 key 对应的时间序列最新样本值加上 addend。
// 如果 key 不存在，则会创建一个新的时间序列，其初始值为 addend，
// 并受 opt 中参数的约束（如保留策略、标签等）。
//
// 返回值为被更新或新增样本的时间戳（毫秒级 Unix 时间戳）。
//
// 参数说明：
//   - key: 时间序列的键名。
//   - addend: 要累加到最新样本的值。
//   - opt: TS.INCRBY 的可选参数，可为 nil。
//
// 返回值：
//   - int64: 更新后样本的时间戳（毫秒）。
func (c *Client) TSIncrBy(ctx context.Context, key string, addend float64, opt *TSIncrByOption) (int64, error) {
	args := []any{"TS.INCRBY", key, addend}
	if opt != nil {
		args = opt.appendArgs(args)
	}
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

type TSIncrByOption = TSDecrByOption

// TSInfo 用于获取时间序列的元数据信息。
//
// 该方法返回 key 对应时间序列的详细信息，例如：
// 保留时间（retention）、样本数量、时间范围、标签、
// 以及压缩规则等元数据。
//
// 当 debug 为 true 时，会返回额外的调试信息（如内部结构或统计信息）。
//
// 参数说明：
//   - ctx: 用于控制取消和超时的上下文。
//   - key: 时间序列的键名。
//   - debug: 是否启用调试模式，返回更多内部信息。
//
// 返回值：
//   - TSInfo: 时间序列的元数据信息结构体。
//   - 若 key 不存在，会返回错误： ERR TSDB: the key does not exist
func (c *Client) TSInfo(ctx context.Context, key string, debug bool) (TSInfo, error) {
	args := []any{"TS.INFO", key}
	if debug {
		args = append(args, "DEBUG")
	}
	cmd := resp3.NewRequest(resp3.DataTypeMap, args...)
	resp := c.do(ctx, cmd)

	data, err := resp3.ToMap(resp.result, resp.err)
	if err != nil {
		return TSInfo{}, err
	}
	info := &TSInfo{}
	err = info.parser(data)
	return *info, err
}

// TSInfo 表示 TS.INFO 命令的返回结果。
type TSInfo struct {
	TotalSamples      int64              `json:"totalSamples"`      // 样本总数
	MemoryUsage       int64              `json:"memoryUsage"`       // 内存占用（字节）
	FirstTimestamp    int64              `json:"firstTimestamp"`    // 首样本时间戳
	LastTimestamp     int64              `json:"lastTimestamp"`     // 尾样本时间戳
	RetentionTime     int64              `json:"retentionTime"`     // 保留时间（毫秒）
	ChunkCount        int64              `json:"chunkCount"`        // 数据块数量
	ChunkSize         int64              `json:"chunkSize"`         // 数据块大小（字节）
	ChunkType         TSEncodingType     `json:"chunkType"`         // 数据块类型，compressed/uncompressed
	DuplicatePolicy   TSDuplicatePolicy  `json:"duplicatePolicy"`   // 重复策略
	Labels            map[string]string  `json:"labels"`            // 标签集合
	SourceKey         *string            `json:"sourceKey"`         // 源时间序列（若是衍生序列）
	Rules             []TSCompactionRule `json:"rules"`             // 规则集合，key 为目标序列名
	IgnoreMaxTimeDiff int64              `json:"ignoreMaxTimeDiff"` // IGNORE 最大时间差
	IgnoreMaxValDiff  float64            `json:"ignoreMaxValDiff"`  // IGNORE 最大值差
	KeySelfName       string             `json:"keySelfName"`       // 时间序列自身名称
	Chunks            []TSInfoChunk      `json:"chunks"`            // 数据块信息列表
}

func (info *TSInfo) parser(data map[resp3.Element]resp3.Element) error {
	var err error
	for k, v := range data {
		var key string
		key, err = resp3.ToString(k, nil)
		if err != nil {
			return err
		}
		switch key {
		case "totalSamples":
			info.TotalSamples, err = resp3.ToInt64(v, nil)
		case "memoryUsage":
			info.MemoryUsage, err = resp3.ToInt64(v, nil)
		case "firstTimestamp":
			info.FirstTimestamp, err = resp3.ToInt64(v, nil)
		case "lastTimestamp":
			info.LastTimestamp, err = resp3.ToInt64(v, nil)
		case "retentionTime":
			info.RetentionTime, err = resp3.ToInt64(v, nil)
		case "chunkCount":
			info.ChunkCount, err = resp3.ToInt64(v, nil)
		case "chunkSize":
			info.ChunkSize, err = resp3.ToInt64(v, nil)
		case "chunkType":
			var s string
			s, err = resp3.ToString(v, nil)
			if err == nil {
				info.ChunkType = TSEncodingType(strings.ToUpper(s))
			}
		case "duplicatePolicy":
			var s string
			s, err = resp3.ToString(v, nil)
			if err == nil {
				info.DuplicatePolicy = TSDuplicatePolicy(strings.ToUpper(s))
			}
		case "labels":
			info.Labels, err = resp3.ToStringMap(v, nil)
		case "sourceKey":
			info.SourceKey, err = resp3.ToPtrString(v, nil)
		case "rules":
			err = info.parserRules(v)
		case "ignoreMaxTimeDiff":
			info.IgnoreMaxTimeDiff, err = resp3.ToInt64(v, nil)
		case "ignoreMaxValDiff":
			info.IgnoreMaxValDiff, err = resp3.ToFloat64(v, nil)
		case "keySelfName":
			info.KeySelfName, err = resp3.ToString(v, nil)
		case "Chunks":
			if err := info.parserChunks(v); err != nil {
				return err
			}
		default:
			// 未知字段，可以选择忽略
		}
		if err != nil {
			return fmt.Errorf("error for key: %q %w", k, err)
		}
	}
	return err
}

func (info *TSInfo) parserRules(data resp3.Element) error {
	rules, err := resp3.ToMap(data, nil)
	if err != nil {
		return err
	}
	if len(rules) == 0 {
		return nil
	}
	for key, item := range rules {
		ks, err := resp3.ToString(key, nil)
		if err != nil {
			return err
		}
		rule := TSCompactionRule{
			DestKey: ks,
		}
		arr, err := resp3.ToSlice(item, nil)
		if err != nil {
			return err
		}
		if len(arr) > 0 {
			rule.BucketDuration, err = resp3.ToInt64(arr[0], nil)
		}
		if len(arr) > 1 {
			rule.Aggregator, err = resp3.ToString(arr[1], nil)
		}
		if len(arr) > 2 {
			// The alignment (since RedisTimeSeries v1.8)
			rule.AlignTimestamp, err = resp3.ToInt64(arr[2], nil)
		}
		if err != nil {
			return err
		}
		info.Rules = append(info.Rules, rule)
	}
	return nil
}

func (info *TSInfo) parserChunks(data resp3.Element) error {
	chunks, err := resp3.ToSlice(data, nil)
	if err != nil {
		return err
	}

	for _, chunkItem := range chunks {
		chunk := TSInfoChunk{}
		chunkMap, err := resp3.ToMap(chunkItem, nil)
		if err != nil {
			return err
		}
		for k, value := range chunkMap {
			ks, err := resp3.ToString(k, nil)
			if err != nil {
				return err
			}
			switch ks {
			case "startTimestamp":
				chunk.StartTimestamp, err = resp3.ToInt64(value, nil)
			case "endTimestamp":
				chunk.EndTimestamp, err = resp3.ToInt64(value, nil)
			case "samples":
				chunk.Samples, err = resp3.ToInt64(value, nil)
			case "size":
				chunk.Size, err = resp3.ToInt64(value, nil)
			case "bytesPerSample":
				chunk.BytesPerSample, err = resp3.ToFloat64(value, nil)
			}
			if err != nil {
				return fmt.Errorf("key(%q):%w", ks, err)
			}
		}
		info.Chunks = append(info.Chunks, chunk)
	}
	return nil
}

// TSCompactionRule 表示时间序列上的一条压缩（聚合）规则。
type TSCompactionRule struct {
	DestKey        string // 目标时间序列 key
	BucketDuration int64  // 聚合时间桶大小（毫秒）
	Aggregator     string // 聚合类型
	AlignTimestamp int64  // 对齐时间戳
}

// TSInfoChunk 表示每个时间序列块的详细信息。
type TSInfoChunk struct {
	StartTimestamp int64   `json:"startTimestamp"` // 块起始时间戳
	EndTimestamp   int64   `json:"endTimestamp"`   // 块结束时间戳
	Samples        int64   `json:"samples"`        // 块内样本数量
	Size           int64   `json:"size"`           // 块大小（字节）
	BytesPerSample float64 `json:"bytesPerSample"` // 每个样本占用字节数
}

// TSMAdd 批量向一个或多个时间序列中添加样本。
//
// 对应 RedisTimeSeries 的 TS.MADD 命令。
// 每个 TSItem 指定目标 key、时间戳、数值以及可选标签。
//
// 参数：
//   - items: 一个或多个要添加的 TSItem。
//
// 返回值：
//   - 每个样本对应的时间戳，顺序与输入 items 相同。
//     如果某个样本添加失败，其对应位置会返回 Redis 的错误。
func (c *Client) TSMAdd(ctx context.Context, items ...TSItem) ([]int64, error) {
	if len(items) == 0 {
		return nil, errNoValues
	}
	args := []any{"TS.ADD"}
	for _, item := range items {
		args = append(args, item.Key, item.Timestamp, item.Value)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64Slice(resp.result, resp.err, len(items))
}

// TSMGet 根据过滤条件批量获取多个时间序列的最新样本。
//
// 对应 RedisTimeSeries 的 TS.MGET 命令。
// 该命令会根据 filterExpr 匹配时间序列，并返回每个序列的最新数据点。
//
// 参数：
//   - latest: 是否返回最新样本（等价于 TS.MGET 的 LATEST 选项）。
//   - selectedLabels: 指定返回的标签列表；为空表示返回全部标签。
//   - filterExpr: 过滤表达式列表，用于匹配时间序列（如 "region=us", "type!=cache"），不能为空
//
// 返回值：
//   - 匹配的时间序列结果集合，每个元素对应一个时间序列。
func (c *Client) TSMGet(ctx context.Context, filterExpr []string, opt *TSMGetOption) ([]TSMGetResult, error) {
	if len(filterExpr) == 0 {
		return nil, errNoFilterExpr
	}
	args := []any{"TS.MGET"}
	if opt != nil {
		args = opt.appendArgs(args)
	}
	args = append(args, "FILTER")
	args = xslice.Append(args, filterExpr...)

	cmd := resp3.NewRequest(resp3.DataTypeMap, args...)
	resp := c.do(ctx, cmd)
	data, err := resp3.ToMap(resp.result, resp.err)
	if err != nil {
		return nil, err
	}
	results := make([]TSMGetResult, 0, len(data))
	for key, item := range data {
		resultItem := TSMGetResult{}
		itemData, err1 := resp3.ToSlice(item, nil)
		resultItem.Key, err1 = resp3.ToString(key, err1)
		if err1 != nil {
			return nil, err1
		}
		if len(itemData) > 0 {
			labels, err2 := resp3.ToStringMap(itemData[0], nil)
			if err2 != nil {
				return nil, err2
			}
			resultItem.Labels = labels
		}
		if len(itemData) > 1 {
			vs, err3 := resp3.ToSlice(itemData[1], nil)
			if err3 != nil {
				return nil, err3
			}
			if len(vs) == 2 {
				ti := &TSSample{}
				var err4 error
				ti.Timestamp, err4 = resp3.ToInt64(vs[0], nil)
				ti.Value, err4 = resp3.ToFloat64(vs[1], err4)
				if err4 != nil {
					return nil, err4
				}
				resultItem.Sample = ti
			}
		}
		results = append(results, resultItem)
	}
	return results, nil
}

type TSMGetOption struct {
	Latest         bool
	WithLabels     bool     // 是否返回 label，和 SelectedLabels 二选一，且 WithLabels 优先
	SelectedLabels []string // 是否返回指定的标签
}

func (opt *TSMGetOption) appendArgs(args []any) []any {
	if opt.Latest {
		args = append(args, "LATEST")
	}
	if opt.WithLabels {
		args = append(args, "WITHLABELS")
	} else if len(opt.SelectedLabels) > 0 {
		args = append(args, "SELECTED_LABELS")
		args = xslice.Append(args, opt.SelectedLabels...)
	}
	return args
}

type TSMGetResult struct {
	Key    string
	Labels map[string]string
	Sample *TSSample
}

// TSMRange 按时间范围批量查询多个时间序列的样本数据。
//
// 对应 RedisTimeSeries 的 TS.MRANGE 命令。
// 根据过滤表达式匹配时间序列，并在指定时间范围内查询样本。
// 可选进行聚合计算，返回聚合后的结果。
//
// 参数：
//   - fromTimestamp: 起始时间戳，毫秒级 Unix 时间戳，或 "-" 表示最早时间。
//   - toTimestamp: 结束时间戳，毫秒级 Unix 时间戳， 或 "+" 表示最新时间。
//   - aggregator: 聚合函数名称（如 "avg"、"sum"、"min"、"max" 等）,不可为空
//   - bucketDuration: 聚合桶的时间跨度（毫秒），
//   - filterExpr: 标签过滤表达式列表，用于匹配时间序列，例如 "region=us", "type!=cache"。
//   - opt: 可选参数，用于控制标签返回、分组、排序、对齐等行为。
//
// 返回值：
//   - 查询得到的样本列表。
func (c *Client) TSMRange(ctx context.Context, fromTimestamp, toTimestamp string, aggregator string, bucketDuration int64, filterExpr []string, opt *TSMRangeOption) ([]TSMRangeResult, error) {
	return c.doTSMRange(ctx, "TS.MRANGE", fromTimestamp, toTimestamp, aggregator, bucketDuration, filterExpr, opt)
}

// TSMRevRange 按时间范围逆序批量查询多个时间序列的样本数据。
//
// 对应 RedisTimeSeries 的 TS.MREVRANGE 命令。
// 根据过滤表达式匹配时间序列，并在指定时间范围内查询样本。
// 可选进行聚合计算，返回聚合后的结果。
//
// 参数：
//   - fromTimestamp: 起始时间戳，毫秒级 Unix 时间戳，或 "-" 表示最早时间。
//   - toTimestamp: 结束时间戳，毫秒级 Unix 时间戳， 或 "+" 表示最新时间。
//   - aggregator: 聚合函数名称（如 "avg"、"sum"、"min"、"max" 等）,不可为空
//   - bucketDuration: 聚合桶的时间跨度（毫秒），
//   - filterExpr: 标签过滤表达式列表，用于匹配时间序列，例如 "region=us", "type!=cache"。
//   - opt: 可选参数，用于控制标签返回、分组、排序、对齐等行为。
//
// 返回值：
//   - 查询得到的样本列表。
func (c *Client) TSMRevRange(ctx context.Context, fromTimestamp, toTimestamp string, aggregator string, bucketDuration int64, filterExpr []string, opt *TSMRangeOption) ([]TSMRangeResult, error) {
	return c.doTSMRange(ctx, "TS.MREVRANGE", fromTimestamp, toTimestamp, aggregator, bucketDuration, filterExpr, opt)
}

func (c *Client) doTSMRange(ctx context.Context, command string, fromTimestamp, toTimestamp string, aggregator string, bucketDuration int64, filterExpr []string, opt *TSMRangeOption) ([]TSMRangeResult, error) {
	args := []any{command, fromTimestamp, toTimestamp}
	if opt != nil {
		args = opt.appendArgs1(args)
	}
	args = append(args, "AGGREGATION", aggregator, bucketDuration)
	if opt != nil {
		args = opt.appendArgs2(args)
	}
	args = append(args, "FILTER")
	args = xslice.Append(args, filterExpr...)

	if opt != nil {
		args = opt.appendArgs3(args)
	}
	cmd := resp3.NewRequest(resp3.DataTypeMap, args...)
	resp := c.do(ctx, cmd)
	mpData, err := resp3.ToMap(resp.result, resp.err)
	if err != nil {
		return nil, err
	}
	results := make([]TSMRangeResult, 0, len(mpData))
	for k, v := range mpData {
		resultItem := TSMRangeResult{}
		resultItem.Key, err = resp3.ToString(k, nil)
		var arr []resp3.Element
		arr, err = resp3.ToSlice(v, err)
		if err != nil {
			return nil, err
		}
		if len(arr) > 0 {
			resultItem.Labels, err = resp3.ToStringMap(arr[0], nil)
		}
		if len(arr) > 1 {
			resultItem.Metadata, err = resp3.ToStringAnyMap(arr[1], err)
		}
		if len(arr) == 3 { // 没有groupBy
			resultItem.Samples, err = tsParserSamples(arr[2], err)
		} else if len(arr) == 4 { // 有 groupby
			var sourceMap resp3.Map
			sourceMap, err = resp3.ToMap(arr[2], err)
			if err != nil {
				return nil, err
			}
			for sk, sv := range sourceMap {
				var sks string
				sks, err = resp3.ToString(sk, nil)
				if err != nil {
					return nil, err
				}
				if sks == "sources" {
					resultItem.Sources, err = resp3.ToStringSlice(sv, err, 0)
					break
				}
			}
			resultItem.Samples, err = tsParserSamples(arr[3], err)
		}
		if err != nil {
			return nil, err
		}
		results = append(results, resultItem)
	}
	return results, nil
}

func tsParserSamples(e resp3.Element, err error) ([]TSSample, error) {
	arr, err := resp3.ToSlice(e, err)
	if err != nil {
		return nil, err
	}
	result := make([]TSSample, 0, len(arr))
	for _, item := range arr {
		sa, err := resp3.ToSlice(item, nil)
		if err != nil {
			return nil, err
		}
		if len(sa) != 2 {
			return nil, fmt.Errorf("invalid array length: %d", len(sa))
		}
		obj := TSSample{}
		obj.Timestamp, err = resp3.ToInt64(sa[0], err)
		obj.Value, err = resp3.ToFloat64(sa[1], err)
		if err != nil {
			return nil, err
		}
		result = append(result, obj)
	}
	return result, nil
}

type TSMRangeResult struct {
	Key      string
	Labels   map[string]string
	Metadata map[string]any
	Sources  []string // groupBy 的时候返回
	Samples  []TSSample
}

// TSMRangeOption 定义 TS.MRANGE / TS.REVMRANGE 命令的可选参数。
type TSMRangeOption struct {
	// Latest 表示返回最新样本（等价于 LATEST 选项）。
	// 当时间序列中存在未来时间戳的数据点时，启用该选项可返回最新值。
	Latest bool

	// FilterByTS 指定仅返回给定时间戳列表中的样本
	// （等价于 FILTER_BY_TS 选项）。
	FilterByTS []int64

	// FilterByValue 指定按数值范围过滤样本
	// （等价于 FILTER_BY_VALUE min max）。
	FilterByValue *TSFilterByValue

	WithLabels bool

	SelectedLabels []string

	// Count 限制返回的样本数量（等价于 COUNT 选项）。
	Count int64

	// Align 指定聚合对齐方（等价于 ALIGN 选项），
	// 常见取值为 "-"、"+" 或具体时间戳字符串。
	Align string

	// BucketTimestamp 指定聚合结果中桶时间戳的表示方式
	// （等价于 BUCKETTIMESTAMP 选项），
	// 例如 "-"、"+" 或 "~"。
	BucketTimestamp string

	// Empty 表示在聚合查询中是否返回空桶（等价于 EMPTY 选项）。
	Empty bool

	GroupBy string

	Reduce string
}

func (opt *TSMRangeOption) appendArgs1(args []any) []any {
	if opt.Latest {
		args = append(args, "LATEST")
	}
	if len(opt.FilterByTS) > 0 {
		args = append(args, "FILTER_BY_TS")
		args = xslice.Append(args, opt.FilterByTS...)
	}
	if opt.FilterByValue != nil {
		args = append(args, "FILTER_BY_VALUE", opt.FilterByValue.Min, opt.FilterByValue.Max)
	}

	if opt.WithLabels {
		args = append(args, "WITHLABELS")
	} else if len(opt.SelectedLabels) > 0 {
		args = append(args, "SELECTED_LABELS")
		args = xslice.Append(args, opt.SelectedLabels...)
	}

	if opt.Count > 0 {
		args = append(args, "COUNT", opt.Count)
	}
	if opt.Align != "" {
		args = append(args, "ALIGN", opt.Align)
	}
	return args
}

func (opt *TSMRangeOption) appendArgs2(args []any) []any {
	if opt.BucketTimestamp != "" {
		args = append(args, "BUCKETTIMESTAMP", opt.BucketTimestamp)
	}
	if opt.Empty {
		args = append(args, "EMPTY")
	}
	return args
}

func (opt *TSMRangeOption) appendArgs3(args []any) []any {
	if opt.GroupBy != "" {
		args = append(args, "GROUPBY", opt.GroupBy, "REDUCE", opt.Reduce)
	}
	return args
}

var errNoFilterExpr = errors.New("no filterExpr")

// TSQueryIndex 根据标签过滤表达式查询匹配的时间序列 key 列表。
//
// 对应 RedisTimeSeries 的 TS.QUERYINDEX 命令。
// 该命令仅返回满足所有过滤条件的时间序列 key，不返回样本数据。
//
// 参数：
//   - filterExprs: 标签过滤表达式列表，用于匹配时间序列，
//     例如 "region=us", "type!=cache"。
//
// 返回值：
//   - 满足过滤条件的时间序列 key 列表。
func (c *Client) TSQueryIndex(ctx context.Context, filterExprs ...string) ([]string, error) {
	if len(filterExprs) == 0 {
		return nil, errNoFilterExpr
	}
	args := []any{"TS.QUERYINDEX"}
	args = xslice.Append(args, filterExprs...)
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToStringSlice(resp.result, resp.err, 0)
}

// TSRange 按时间范围查询时间序列中的样本数据。
//
// 对应 RedisTimeSeries 的 TS.RANGE 命令。可指定时间范围，并可选进行聚合计算返回聚合后的样本结果。
//
// 参数：
//   - key: 时间序列的 key。
//   - fromTimestamp: 起始时间戳，毫秒级 Unix 时间戳，或 "-" 表示最早时间。
//   - toTimestamp: 结束时间戳，毫秒级 Unix 时间戳，或 "+" 表示最新时间。
//   - aggregator: 聚合函数名称（如 "avg"、"sum"、"min"、"max" 等），不可为空。
//   - bucketDuration: 聚合桶的时间跨度（毫秒）。
//   - opt: 可选参数，用于控制对齐方式、过滤条件、返回数量等。
//
// 返回值：
//   - 查询得到的样本列表，按时间戳升序排列。
func (c *Client) TSRange(ctx context.Context, key string, fromTimestamp, toTimestamp string, aggregator string, bucketDuration int64, opt *TSRangeOption) ([]TSSample, error) {
	return c.doTSRange(ctx, "TS.RANGE", key, fromTimestamp, toTimestamp, aggregator, bucketDuration, opt)
}

// TSRevRange 按时间范围逆序查询时间序列中的样本数据。
//
// 对应 RedisTimeSeries 的 TS.REVRANGE 命令。
// 功能与 TSRange 基本一致，但结果按时间戳降序返回。
// 可指定时间范围，并可选进行聚合计算返回聚合后的样本结果。
//
// 参数：
//   - key: 时间序列的 key。
//   - fromTimestamp: 起始时间戳，毫秒级 Unix 时间戳，或 "-" 表示最早时间。
//   - toTimestamp: 结束时间戳，毫秒级 Unix 时间戳，或 "+" 表示最新时间。
//   - aggregator: 聚合函数名称（如 "avg"、"sum"、"min"、"max" 等）,不可为空
//   - bucketDuration: 聚合桶的时间跨度（毫秒），
//   - opt: 可选参数，用于控制对齐方式、过滤条件、返回数量等。
//
// 返回值：
//   - 查询得到的样本列表，按时间戳降序排列。
func (c *Client) TSRevRange(ctx context.Context, key string, fromTimestamp, toTimestamp string, aggregator string, bucketDuration int64, opt *TSRangeOption) ([]TSSample, error) {
	return c.doTSRange(ctx, "TS.REVRANGE", key, fromTimestamp, toTimestamp, aggregator, bucketDuration, opt)
}

func (c *Client) doTSRange(ctx context.Context, command string, key string, fromTimestamp, toTimestamp string, aggregator string, bucketDuration int64, opt *TSRangeOption) ([]TSSample, error) {
	args := []any{command, key, fromTimestamp, toTimestamp}
	if opt != nil {
		args = opt.appendArgs1(args)
	}
	args = append(args, "AGGREGATION", aggregator, bucketDuration)
	if opt != nil {
		args = opt.appendArgs2(args)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	arr, err := resp.asResp3Array(0)
	if err != nil {
		return nil, err
	}
	result := make([]TSSample, 0, len(arr))
	for _, item := range arr {
		sa, err2 := resp3.ToSlice(item, nil)
		if err2 != nil {
			return nil, err2
		}
		if len(sa) != 2 {
			return nil, fmt.Errorf("invalid array length: %d", len(sa))
		}
		obj := TSSample{}
		obj.Timestamp, err = resp3.ToInt64(sa[0], err)
		obj.Value, err = resp3.ToFloat64(sa[1], err)
		if err != nil {
			return nil, err
		}
		result = append(result, obj)
	}
	return result, nil
}

// TSRangeOption 定义 TS.RANGE / TS.REVRANGE 命令的可选参数。
type TSRangeOption struct {
	// Latest 表示返回最新样本（等价于 LATEST 选项）。
	// 当时间序列中存在未来时间戳的数据点时，启用该选项可返回最新值。
	Latest bool

	// FilterByTS 指定仅返回给定时间戳列表中的样本
	// （等价于 FILTER_BY_TS 选项）。
	FilterByTS []int64

	// FilterByValue 指定按数值范围过滤样本
	// （等价于 FILTER_BY_VALUE min max）。
	FilterByValue *TSFilterByValue

	// Count 限制返回的样本数量（等价于 COUNT 选项）。
	Count int64

	// Align 指定聚合对齐方（等价于 ALIGN 选项），
	// 常见取值为 "-"、"+" 或具体时间戳字符串。
	Align string

	// BucketTimestamp 指定聚合结果中桶时间戳的表示方式
	// （等价于 BUCKETTIMESTAMP 选项），
	// 例如 "-"、"+" 或 "~"。
	BucketTimestamp string

	// Empty 表示在聚合查询中是否返回空桶（等价于 EMPTY 选项）。
	Empty bool
}

func (opt *TSRangeOption) appendArgs1(args []any) []any {
	if opt.Latest {
		args = append(args, "LATEST")
	}
	if len(opt.FilterByTS) > 0 {
		args = append(args, "FILTER_BY_TS")
		args = xslice.Append(args, opt.FilterByTS...)
	}
	if opt.FilterByValue != nil {
		args = append(args, "FILTER_BY_VALUE", opt.FilterByValue.Min, opt.FilterByValue.Max)
	}
	if opt.Count > 0 {
		args = append(args, "COUNT", opt.Count)
	}
	if opt.Align != "" {
		args = append(args, "ALIGN", opt.Align)
	}
	return args
}

func (opt *TSRangeOption) appendArgs2(args []any) []any {
	if opt.BucketTimestamp != "" {
		args = append(args, "BUCKETTIMESTAMP", opt.BucketTimestamp)
	}
	if opt.Empty {
		args = append(args, "EMPTY")
	}
	return args
}

// TSFilterByValue 定义按数值范围过滤样本的条件。
//
// 对应 RedisTimeSeries 的 FILTER_BY_VALUE min max 选项。
// 仅返回数值在 [Min, Max] 区间内的样本。
type TSFilterByValue struct {
	// Min 为数值下限（包含）。
	Min float64

	// Max 为数值上限（包含）。
	Max float64
}
