//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-12-25

package xredis

import (
	"context"
	"fmt"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/store/xredis/resp3"
)

// https://redis.io/docs/latest/commands/bitcount/

// BitCount 返回存储在指定 key 的字符串值中，值为 1 的比特位总数。
//
// 该方法对应 Redis 的 BITCOUNT 命令，用于统计整个字符串中被置为 1 的比特数。
//
// 参数：
//   - key: Redis 中要统计的 key
//
// 返回值：
//   - int64: key 对应字符串中比特值为 1 的总数
func (c *Client) BitCount(ctx context.Context, key string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "BITCOUNT", key)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// BitCountRange 返回存储在指定 key 的字符串值中，指定范围内值为 1 的比特位总数。
//
// 该方法对应 Redis 的 BITCOUNT 命令带范围选项，用于统计字符串中
// 某个区间内的比特位为 1 的数量。
//
// 参数：
//   - key: Redis 中要统计的 key
//   - start: 起始偏移量（可以为负数，表示从字符串末尾开始计数）
//   - end: 结束偏移量（可以为负数）
//   - unit: 统计单位，可选值：""（空字符串） 或 "BYTE"（按字节）或 "BIT"（按位）
//
// 返回值：
//   - int64: 指定范围内比特值为 1 的总数
//   - error: 操作过程中可能产生的错误
//
// 示例：
//
//	// 统计 key "mykey" 从第 0 字节到第 10 字节的 1 的总数
//	count, err := client.BitCountRange(ctx, "mykey", 0, 10, "BYTE")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("1 的总数:", count)
func (c *Client) BitCountRange(ctx context.Context, key string, start int64, end int64, unit string) (int64, error) {
	args := []any{"BITCOUNT", key, start, end}
	if unit != "" {
		args = append(args, unit)
	}
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// BitField 对存储在指定 key 的字符串值执行 BITFIELD 命令操作。
//
// 该方法可以同时执行多个子命令（GET / SET / INCRBY / OVERFLOW），
// 并返回每个操作的结果。使用可变参数 ops 传入不同的操作选项。
//
// 参数：
//   - ctx: 上下文，用于控制超时或取消操作
//   - key: Redis 中要操作的 key
//   - ops: 可变参数，传入具体的 BitFieldOption 操作（如 BitFieldGet、BitFieldSet、BitFieldIncrBy、BitFieldOverflow）
//
// 返回值：
//   - []*int64: 每个操作对应的结果列表，顺序与 ops 一致。
//   - 对于 GET 和 INCRBY 操作，返回对应的整数值或 nil（当 INCRBY 配合 FAIL 且溢出时）
//   - 对于 SET 操作，返回设置前的旧值
func (c *Client) BitField(ctx context.Context, key string, ops ...BitFieldOption) ([]*int64, error) {
	args := []any{"BITFIELD", key}

	for _, op := range ops {
		args = op.appendArgs(args)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToPtrInt64Slice(resp.result, resp.err)
}

// BitFieldOption BitField 方法的配置参数
type BitFieldOption interface {
	bitFieldOpRo() bool
	appendArgs(args []any) []any
}

// BitFieldGet 表示 Redis BITFIELD 命令中的 GET 子命令参数。
//
// 该结构体用于描述如何从 Redis 字符串值中读取一段位字段（bit field），
// 通过 Encoding 指定位字段的类型和长度，通过 Offset 指定读取的起始位偏移。
type BitFieldGet struct {
	// Encoding 用于指定位字段的编码方式。
	//
	// 编码格式为：
	//
	//	[i|u]<bits>
	//
	// 其中：
	//   - i 表示有符号整数（使用二进制补码表示）
	//   - u 表示无符号整数
	//   - <bits> 表示位宽，取值范围为 [1, 64]
	//
	// 常见示例：
	//   "u8"   // 8 位无符号整数（0 ~ 255）
	//   "i8"   // 8 位有符号整数（-128 ~ 127）
	//   "u16"  // 16 位无符号整数
	//   "i32"  // 32 位有符号整数
	//
	// Encoding 决定了从 Offset 开始连续读取多少位，
	// 以及这些位如何被解析为最终的整数值。
	Encoding string

	// Offset 表示读取位字段的起始偏移量（从 0 开始）。
	//
	// 该偏移量以「位（bit）」为单位，而不是字节（byte）, 因此不要求按字节对齐。
	//
	// Redis 会从 Offset 指定的位置开始，连续读取 Encoding 所定义的位宽数量的比特。
	//
	// 在 GET 操作中，如果读取范围超出了当前字符串长度，超出的比特将被视为 0，不会修改原有的 key。
	Offset int64
}

func (BitFieldGet) bitFieldOpRo() bool {
	return true
}

func (bg BitFieldGet) appendArgs(args []any) []any {
	return append(args, "GET", bg.Encoding, bg.Offset)
}

// BitFieldSet 表示 Redis BITFIELD 命令中的 SET 子命令参数。
//
// 该结构体用于描述如何向 Redis 字符串值中写入一段位字段（bit field），
// 通过 Encoding 指定位字段的类型和长度，通过 Offset 指定写入的起始位偏移，并将 Value 按指定编码写入对应的位范围。
type BitFieldSet struct {
	// Encoding 用于指定位字段的编码方式。
	//
	// 编码格式为：
	//
	//	[i|u]<bits>
	//
	// 其中：
	//   - i 表示有符号整数（使用二进制补码表示）
	//   - u 表示无符号整数
	//   - <bits> 表示位宽，取值范围为 [1, 64]
	//
	// 常见示例：
	//   "u8"   // 8 位无符号整数（0 ~ 255）
	//   "i8"   // 8 位有符号整数（-128 ~ 127）
	//   "u16"  // 16 位无符号整数
	//   "i32"  // 32 位有符号整数
	//
	// Encoding 决定了从 Offset 开始连续读取多少位，
	// 以及这些位如何被解析为最终的整数值。
	Encoding string

	// Offset 表示写入位字段的起始偏移量（从 0 开始）。
	//
	// 偏移量以「位（bit）」为单位，而非字节（byte），
	// 不要求按字节对齐。
	//
	// Redis 会从 Offset 开始，连续写入 Encoding 所定义的位宽。
	// 如果写入范围超出当前字符串长度，Redis 会自动扩展字符串，
	// 并将未覆盖的比特填充为 0。
	Offset int64

	// Value 表示要写入位字段的整数值。
	//
	// 该值会按照 Encoding 指定的类型（有符号 / 无符号）和位宽进行截断或转换后写入。
	//
	// 对于有符号编码（i<bits>），Value 使用二进制补码形式存储；
	// 对于无符号编码（u<bits>），Value 应为非负整数。
	Value int64
}

func (BitFieldSet) bitFieldOpRo() bool {
	return false
}

func (bs BitFieldSet) appendArgs(args []any) []any {
	return append(args, "SET", bs.Encoding, bs.Offset, bs.Value)
}

// BitFieldIncrBy 表示 Redis BITFIELD 命令中的 INCRBY 子命令参数。
//
// 该结构体用于描述对某一段位字段执行自增（或自减）操作，
// Redis 会按照指定的 Encoding 从 Offset 位置读取位字段，
// 在其基础上加上 Increment，并将结果写回原位置。
type BitFieldIncrBy struct {
	// Encoding 用于指定位字段的编码方式。
	//
	// 编码格式为：
	//
	//	[i|u]<bits>
	//
	// 其中：
	//   - i 表示有符号整数（使用二进制补码表示）
	//   - u 表示无符号整数
	//   - <bits> 表示位宽，取值范围为 [1, 64]
	//
	// 常见示例：
	//   "u8"   // 8 位无符号整数（0 ~ 255）
	//   "i8"   // 8 位有符号整数（-128 ~ 127）
	//   "u16"  // 16 位无符号整数
	//   "i32"  // 32 位有符号整数
	//
	// Encoding 决定了从 Offset 开始连续读取多少位，
	// 以及这些位如何被解析为最终的整数值。
	Encoding string

	// Offset 表示位字段的起始偏移量（从 0 开始）。
	//
	// 偏移量以「位（bit）」为单位，而不是字节（byte），
	// 不要求按字节对齐。
	//
	// Redis 会从 Offset 开始读取 Encoding 指定宽度的位字段，
	// 并在计算完成后将结果写回同一位置。
	//
	// 如果操作范围超出当前字符串长度，
	// Redis 会自动扩展字符串并用 0 填充缺失的比特。
	Offset int64

	// Increment 表示要增加（或减少）的整数值。
	//
	// 该值会在当前位字段值的基础上相加：
	//
	//	newValue = oldValue + Increment
	//
	// Increment 可以为负数，用于实现递减操作。
	//
	// 当计算结果超出 Encoding 所定义的取值范围时，
	// 实际行为取决于 BITFIELD 命令是否指定了 OVERFLOW 子句：
	//   - WRAP（默认）：按位宽回绕
	//   - SAT：结果被限制在最小/最大值
	//   - FAIL：操作失败并返回 nil
	Increment int64
}

func (BitFieldIncrBy) bitFieldOpRo() bool {
	return false
}

func (bi BitFieldIncrBy) appendArgs(args []any) []any {
	return append(args, "INCRBY", bi.Encoding, bi.Offset, bi.Increment)
}

// BitFieldOverflow 表示 Redis BITFIELD 命令中 INCRBY 子命令的溢出处理策略。
//
//   - 通过布尔值标记选择哪种溢出策略。 BITFIELD 命令中一次只能选择一种策略，
//   - 如果多个字段同时为 true，按照 WRAP、SAT、FAIL 先后顺序，先为 true 的被采用。
//   - 若所有字段同时为 false，则忽略该配置，采用 Redis Server 默认的策略（一般是 WRAP）。
type BitFieldOverflow struct {
	// WRAP 表示溢出时按位宽回绕（wrap）。
	//
	// 例如对于 8 位无符号整数：
	//   255 + 1 → 0
	WRAP bool

	// SAT 表示溢出时饱和（saturate）。
	//
	// 当计算结果超出范围时，将写入最小或最大值：
	//   8 位无符号整数：0 或 255
	//   8 位有符号整数：-128 或 127
	SAT bool

	// FAIL 表示溢出时操作失败。
	//
	// 当计算结果超出 Encoding 定义的取值范围时，
	// 不写入任何值，INCRBY 返回 nil。
	FAIL bool
}

func (BitFieldOverflow) bitFieldOpRo() bool {
	return false
}

func (bo BitFieldOverflow) appendArgs(args []any) []any {
	if bo.WRAP {
		return append(args, "OVERFLOW", "WRAP")
	} else if bo.SAT {
		return append(args, "OVERFLOW", "SAT")
	} else if bo.FAIL {
		return append(args, "OVERFLOW", "FAIL")
	}
	return args
}

// BitFieldRo 对存储在指定 key 的字符串值执行只读 BITFIELD 命令操作。
//
// 该方法仅支持不修改 key 的操作（如 GET），与 BitField 方法类似，但不会对数据进行写入。
// 返回值为每个操作的结果，顺序与传入的 ops 一致。
//
// 参数：
//   - key: Redis 中要操作的 key
//   - ops: 可变参数，传入具体的 BitFieldGet 操作（仅支持 GET 操作或不修改数据的操作）
//
// 返回值：
//   - []int64: 每个操作对应的结果列表
func (c *Client) BitFieldRo(ctx context.Context, key string, ops ...BitFieldOption) ([]int64, error) {
	args := []any{"BITFIELD_RO", key}

	for _, op := range ops {
		if !op.bitFieldOpRo() {
			return nil, fmt.Errorf("has write operation: %v", op.appendArgs(nil))
		}
		args = op.appendArgs(args)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64Slice(resp.result, resp.err)
}

// BitOP 对一个或多个 key 的字符串值执行按位运算，并将结果存储到目标 key 中。
//
// 该方法对应 Redis 的 BITOP 命令，支持 AND、OR、XOR 和 NOT 等按位运算。
// 当操作类型为 NOT 时，只能指定一个 source key。
//
// 参数：
//   - op: 按位运算类型，可选值： AND | OR | XOR | NOT | DIFF | DIFF1 | ANDOR | ONE
//   - destKey: 存储结果的 key
//   - keys: 操作的 key 列表,必填
//
// 返回值：
//   - int64: 目标 key 的字符串长度（以字节为单位）
func (c *Client) BitOP(ctx context.Context, op string, destKey string, keys ...string) (int64, error) {
	if len(keys) == 0 {
		return 0, errNoKeys
	}
	args := []any{"BITOP", op, destKey}
	args = xslice.Append(args, keys...)
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// BitPos 返回指定 key 中第一个值为 bit 的比特位的位置（以 bit 为单位）。
//
// 该方法对应 Redis 的 BITPOS 命令，可以选择可选范围和单位。
// 返回值以 bit 偏移量表示，从 0 开始计数。如果 key 不存在或范围内没有找到目标比特，返回 -1。
//
// 参数：
//
//   - key: Redis 中要操作的 key
//   - bit: 要查找的比特值，只能是 0 或 1
//   - extArgs: 可选参数列表(最多支持 3 个值)，按顺序传入：
//     1. start (int)：起始字节索引（包含）
//     2. end (int)：结束字节索引（包含）
//     3. unit (string)：统计单位，可选 "BYTE"（默认）或 "BIT"
//
// 返回值：
//   - int64: 第一个匹配比特的偏移量，如果未找到则返回 -1
//   - 若 key 不存在，查找 bit = 1 时会返回 -1，nil，查找 bit = 0 时会返回 0,nil
func (c *Client) BitPos(ctx context.Context, key string, bit int, extArgs ...any) (int64, error) {
	if bit != 0 && bit != 1 {
		return 0, fmt.Errorf("bit value %d is invalid: only 0 or 1 are allowed", bit)
	}
	args := []any{"BITPOS", key, bit}

	if argLen := len(extArgs); argLen > 0 {
		if argLen > 3 {
			return 0, fmt.Errorf("extArgs length %d is invalid, max allow 3", argLen)
		}
		if start, ok := extArgs[0].(int); ok {
			args = append(args, start)
		} else {
			return 0, fmt.Errorf("extArgs[0] must be int: %#v", extArgs[0])
		}

		if argLen > 1 {
			if end, ok := extArgs[1].(int); ok {
				args = append(args, end)
			} else {
				return 0, fmt.Errorf("extArgs[1] must be int: %#v", extArgs[1])
			}
		}

		if argLen > 2 {
			if unit, ok := extArgs[2].(string); ok {
				args = append(args, unit)
			} else {
				return 0, fmt.Errorf("extArgs[2] must be string: %#v", extArgs[2])
			}
		}
	}

	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// GetBit 返回存储在指定 key 的字符串值中，指定偏移量的比特位的值。
//
// 该方法对应 Redis 的 GETBIT 命令，用于获取字符串中某一位的值。
// 如果 key 不存在，Redis 会返回 0。
//
// 参数：
//   - key: Redis 中要操作的 key
//   - offset: 要读取的比特位的偏移量（从 0 开始，单位为 bit）
//
// 返回值：
//   - int: 指定偏移量的比特值（0 或 1）
func (c *Client) GetBit(ctx context.Context, key string, offset int64) (int, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "GETBIT", key, offset)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

// BitGet GetBit 的别名
func (c *Client) BitGet(ctx context.Context, key string, offset int64) (int, error) {
	return c.GetBit(ctx, key, offset)
}

// SetBit 将存储在指定 key 的字符串值中，指定偏移量的比特位设置为指定值。
//
// 该方法对应 Redis 的 SETBIT 命令，用于修改字符串中某一位的值。
// 如果 key 不存在，Redis 会自动创建一个新的字符串。
//
// 参数：
//   - key: Redis 中要操作的 key
//   - offset: 要设置的比特位的偏移量（从 0 开始）
//   - bit: 要设置的值，只能为 0 或 1
//
// 返回值：
//   - int: 指定偏移量原来的比特值（0 或 1）
func (c *Client) SetBit(ctx context.Context, key string, offset int64, bit int) (int, error) {
	if bit != 0 && bit != 1 {
		return 0, fmt.Errorf("bit value %d is invalid: only 0 or 1 are allowed", bit)
	}
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "SETBIT", key, offset, bit)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

// BitSet SetBit 方法的别名
func (c *Client) BitSet(ctx context.Context, key string, offset int64, bit int) (int, error) {
	return c.SetBit(ctx, key, offset, bit)
}
