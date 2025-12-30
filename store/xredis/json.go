//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-12-29

package xredis

// https://redis.io/docs/latest/commands/json.arrappend/

import (
	"context"
	"fmt"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/store/xredis/resp3"
)

// JSONArrAppend 用于向指定 key 的 JSON 文档中，目标路径指向的数组末尾追加一个或多个元素。
//
// 等价于 RedisJSON 命令：JSON.ARRAPPEND key [path] value [value ...]
//
// 参数说明：
//   - key: Redis 键名。
//   - path: JSONPath 表达式，必须指向一个 JSON 数组。可选，不为空时有效。
//   - values: 要追加的值列表，必须是可序列化为合法 JSON 的数据。
//
// 若 key 不存在，会返回错误：ERR could not perform this operation on a key that doesn't exist
func (c *Client) JSONArrAppend(ctx context.Context, key string, path string, values ...any) ([]*int64, error) {
	if len(values) == 0 {
		return nil, errNoValues
	}
	args := []any{"JSON.ARRAPPEND", key}
	if path != "" {
		args = append(args, path)
	}
	args = append(args, values...)
	cmd := resp3.NewRequest(resp3.DataTypeAny, args...)
	resp := c.do(ctx, cmd)
	return parserJSONPtrInt64Slice(resp.result, resp.err)
}

func (c *Client) JSONArrIndex(ctx context.Context, key string, path string, value any) ([]*int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeAny, "JSON.ARRINDEX", key, path, value)
	resp := c.do(ctx, cmd)
	return parserJSONPtrInt64Slice(resp.result, resp.err)
}

func (c *Client) JSONArrIndexRange(ctx context.Context, key string, path string, value any, start int, stop int) ([]*int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeAny, "JSON.ARRINDEX", key, path, value, start, stop)
	resp := c.do(ctx, cmd)
	return parserJSONPtrInt64Slice(resp.result, resp.err)
}

func (c *Client) JSONArrInsert(ctx context.Context, key string, path string, index int, values ...any) ([]*int64, error) {
	if len(values) == 0 {
		return nil, errNoValues
	}
	args := []any{"JSON.ARRINSERT", key, path, index}
	args = xslice.Append(args, values...)
	cmd := resp3.NewRequest(resp3.DataTypeAny, args...)
	resp := c.do(ctx, cmd)
	return parserJSONPtrInt64Slice(resp.result, resp.err)
}

func (c *Client) JSONArrLen(ctx context.Context, key string, path string) ([]*int64, error) {
	var args []any
	if path == "" {
		args = []any{"JSON.ARRLEN", key}
	} else {
		args = []any{"JSON.ARRLEN", key, path}
	}
	cmd := resp3.NewRequest(resp3.DataTypeAny, args...)
	resp := c.do(ctx, cmd)
	return parserJSONPtrInt64Slice(resp.result, resp.err)
}

func (c *Client) JSONArrPop(ctx context.Context, key string, opt *JSONArrPopOption) ([]*string, error) {
	args := []any{"JSON.ARRPOP", key}
	if opt != nil {
		args = opt.appendArgs(args)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToPtrStringSlice(resp.result, resp.err, 0)
}

type JSONArrPopOption struct {
	Path  string
	Index *int
}

func (opt *JSONArrPopOption) appendArgs(args []any) []any {
	args = append(args, opt.Path)
	if opt.Index != nil {
		args = append(args, *opt.Index)
	}
	return args
}

func (c *Client) JSONArrTrim(ctx context.Context, key string, path string, start int, stop int) ([]*int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeAny, "JSON.ARRTRIM", key, path, start, stop)
	resp := c.do(ctx, cmd)
	return parserJSONPtrInt64Slice(resp.result, resp.err)
}

func (c *Client) JSONClear(ctx context.Context, key string) (int, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "JSON.CLEAR", key)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

func (c *Client) JSONClearPath(ctx context.Context, key string, path string) (int, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "JSON.CLEAR", key, path)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

func (c *Client) JSONDebugMemory(ctx context.Context, key string, path string) (int64, error) {
	var args []any
	if path == "" {
		args = []any{"JSON.DEBUG", "MEMORY", key}
	} else {
		args = []any{"JSON.DEBUG", "MEMORY", key, path}
	}
	cmd := resp3.NewRequest(resp3.DataTypeInteger, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

func (c *Client) JSONDel(ctx context.Context, key string) (int, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "JSON.DEL", key)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

func (c *Client) JSONDelPath(ctx context.Context, key string, path string) (int, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "JSON.DEL", key, path)
	resp := c.do(ctx, cmd)
	return resp3.ToInt(resp.result, resp.err)
}

// JSONGet 用于从指定 key 的 JSON 文档中获取一个或多个路径对应的值。
//
// 等价于 RedisJSON 命令：JSON.GET key [path ...]
//
// 参数说明：
//   - key: Redis 键名。
//   - paths: 可选的 JSONPath 表达式。
//   - 未指定时，默认等价于使用 "$"，返回整个 JSON 文档。
//   - 指定多个 path 时，会一次性返回多个路径对应的结果。
func (c *Client) JSONGet(ctx context.Context, key string, paths ...string) (string, error) {
	args := []any{"JSON.GET", key}
	args = xslice.Append(args, paths...)
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToString(resp.result, resp.err)
}

func (c *Client) JSONGetWithOption(ctx context.Context, key string, opt *JSONGetOption, paths ...string) (string, error) {
	args := []any{"JSON.GET", key}
	if opt != nil {
		args = opt.appendArgs(args)
	}
	args = xslice.Append(args, paths...)
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToString(resp.result, resp.err)
}

type JSONGetOption struct {
	Indent  string
	NewLine string
	Space   string
}

func (opt *JSONGetOption) appendArgs(args []any) []any {
	if opt.Indent != "" {
		args = append(args, "INDENT", opt.Indent)
	}
	if opt.NewLine != "" {
		args = append(args, "NEWLINE", opt.NewLine)
	}

	if opt.Space != "" {
		args = append(args, "SPACE", opt.Space)
	}

	return args
}

func (c *Client) JSONMerge(ctx context.Context, key string, path string, value any) error {
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, "JSON.MERGE", key, path, value)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

func (c *Client) JSONMGet(ctx context.Context, path string, keys ...string) ([]*string, error) {
	if len(keys) == 0 {
		return nil, errNoKeys
	}
	args := []any{"JSON.MGET"}
	args = xslice.Append(args, keys...)
	args = append(args, path)
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToPtrStringSlice(resp.result, resp.err, 0)
}

func (c *Client) JSONMSet(ctx context.Context, items ...JSONItem) error {
	if len(items) == 0 {
		return errNoValues
	}
	args := []any{"JSON.MSET"}
	for _, item := range items {
		args = append(args, item.Key, item.Path, item.Value)
	}
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

type JSONItem struct {
	Key   string
	Path  string
	Value any
}

func (c *Client) JSONNumIncrBy(ctx context.Context, key string, path string, num any) ([]any, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "JSON.NUMINCRBY", key, path, num)
	resp := c.do(ctx, cmd)
	arr, err := resp3.ToSlice(resp.result, resp.err)
	if err != nil || len(arr) == 0 {
		return nil, err
	}
	result := make([]any, len(arr))
	for i, v := range arr {
		switch rv := v.(type) {
		case resp3.Integer:
			result[i] = rv.Int64()
		case resp3.Double:
			result[i] = rv.Float64()
		case resp3.Null:
			result[i] = nil
		default:
			return nil, fmt.Errorf("invalid reply: %#v", v)
		}
	}
	return result, nil
}

func (c *Client) JSONNumMultBy(ctx context.Context, key string, path string, num any) ([]any, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "JSON.NUMMULTBY", key, path, num)
	resp := c.do(ctx, cmd)
	arr, err := resp3.ToSlice(resp.result, resp.err)
	if err != nil || len(arr) == 0 {
		return nil, err
	}
	result := make([]any, len(arr))
	for i, v := range arr {
		switch rv := v.(type) {
		case resp3.Integer:
			result[i] = rv.Int64()
		case resp3.Double:
			result[i] = rv.Float64()
		case resp3.Null:
			result[i] = nil
		default:
			return nil, fmt.Errorf("invalid reply: %#v", v)
		}
	}
	return result, nil
}

func (c *Client) JSONObjKeys(ctx context.Context, key string, path string) ([][]string, error) {
	var args []any
	if path == "" {
		args = []any{"JSON.OBJKEYS", key}
	} else {
		args = []any{"JSON.OBJKEYS", key, path}
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	arr, err := resp3.ToSlice(resp.result, resp.err)
	if err != nil || len(arr) == 0 {
		return nil, err
	}
	result := make([][]string, 0, len(arr))
	for _, v := range arr {
		if v.DataType() == resp3.DataTypeNull {
			result = append(result, nil)
			continue
		}
		keys, err := resp3.ToStringSlice(v, nil, 0)
		if err != nil {
			return nil, err
		}
		result = append(result, keys)
	}
	return result, nil
}

func (c *Client) JSONObjLen(ctx context.Context, key string, path string) ([]*int64, error) {
	var args []any
	if path == "" {
		args = []any{"JSON.OBJLEN", key}
	} else {
		args = []any{"JSON.OBJLEN", key, path}
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return parserJSONPtrInt64Slice(resp.result, resp.err)
}

// JSONSet 用于在指定 key 的 JSON 文档中设置或更新指定路径上的值。
//
// 等价于 RedisJSON 命令：JSON.SET key path value
//
// 参数说明：
//   - key: Redis 键名。
//   - path: JSONPath 表达式，用于指定要设置的 JSON 节点位置。
//   - value: 要写入的值，必须是可序列化为合法 JSON 的数据。
//
// 行为说明：
//   - 当 path 为 "$" 时，整个 JSON 文档会被 value 覆盖。
//   - 当 path 指向已存在节点时，会更新该节点的值。
//   - 当 path 不存在时，默认会创建对应路径（具体行为取决于 RedisJSON 版本）。
func (c *Client) JSONSet(ctx context.Context, key string, path string, value any) error {
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, "JSON.SET", key, path, value)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

// JSONSetNX 用于在指定 key 的 JSON 文档中，仅当目标路径不存在时设置值。
//
// 等价于 RedisJSON 命令：JSON.SET key path value NX
//
// 参数说明：
//   - key: Redis 键名。
//   - path: JSONPath 表达式，用于指定要设置的 JSON 节点位置。
//   - value: 要写入的值，必须是可序列化为合法 JSON 的数据。
//
// 行为说明：
//   - 仅当 path 不存在时才会写入 value。
//   - 如果 path 已存在，则不会修改原有数据。
//   - 不会自动覆盖已有 JSON 节点。
//
// 返回值：
//   - bool: 当成功写入时返回 true；如果 path 已存在而未写入，则返回 false。
//   - error: 当命令执行失败或 value 无法序列化为 JSON 时返回错误。
func (c *Client) JSONSetNX(ctx context.Context, key string, path string, value any) (bool, error) {
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, "JSON.SET", key, path, value, "NX")
	resp := c.do(ctx, cmd)
	return resp3.ToOkBool(resp.result, resp.err)
}

// JSONSetXX 用于在指定 key 的 JSON 文档中，仅当目标路径已存在时更新值。
//
// 等价于 RedisJSON 命令：JSON.SET key path value XX
//
// 参数说明：
//   - ctx: 上下文，用于控制请求的生命周期。
//   - key: Redis 键名。
//   - path: JSONPath 表达式，用于指定要更新的 JSON 节点位置。
//   - value: 要写入的值，必须是可序列化为合法 JSON 的数据。
//
// 行为说明：
//   - 仅当 path 已存在时才会更新 value。
//   - 如果 path 不存在，则不会创建新的节点，也不会修改数据。
//   - 可用于保证不会意外创建新的 JSON 节点。
//
// 返回值：
//   - bool: 当成功更新时返回 true；如果 path 不存在而未更新，则返回 false。
//   - error: 当命令执行失败或 value 无法序列化为 JSON 时返回错误。
func (c *Client) JSONSetXX(ctx context.Context, key string, path string, value any) (bool, error) {
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, "JSON.SET", key, path, value, "XX")
	resp := c.do(ctx, cmd)
	return resp3.ToOkBool(resp.result, resp.err)
}

func (c *Client) JSONStrAppend(ctx context.Context, key string, path string, value string) ([]*int64, error) {
	var args []any
	if path == "" {
		args = []any{"JSON.STRAPPEND", key, value}
	} else {
		args = []any{"JSON.STRAPPEND", key, path, value}
	}
	cmd := resp3.NewRequest(resp3.DataTypeAny, args...)
	resp := c.do(ctx, cmd)
	return parserJSONPtrInt64Slice(resp.result, resp.err)
}

func parserJSONPtrInt64Slice(data resp3.Element, err error) ([]*int64, error) {
	if err != nil {
		return nil, err
	}
	switch rv := data.(type) {
	case resp3.Integer:
		num := rv.Int64()
		return []*int64{&num}, nil
	case resp3.Null:
		// 类型不对
		return []*int64{nil}, nil
	case resp3.Array:
		return resp3.ToPtrInt64Slice(rv, nil)
	case resp3.SimpleError:
		return nil, rv
	default:
		return nil, fmt.Errorf("invalid type for result: %T", rv)
	}
}

// JSONStrLen 返回指定 key 对应 JSON 文本的长度（字符数）。
//
// key 必须对应一个 JSON 字符串值，否则会返回错误。
//
// 返回值：
//   - int64：字符串的长度（字符数）。
//   - 若 key 不存在，会返回 0，ErrNil
func (c *Client) JSONStrLen(ctx context.Context, key string) (int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeInteger, "JSON.STRLEN", key)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64(resp.result, resp.err)
}

// JSONStrLenWithPath 返回指定 key 对应 JSON 文本的长度（字符数）。
//
// key 必须对应一个 JSON 字符串值，否则会返回错误。
//
// 返回值：
//   - int64：字符串的长度（字符数）,如果匹配的值不是字符串类型，返回 nil
//   - 若 key 不存在，会返回 0，ErrNil
func (c *Client) JSONStrLenWithPath(ctx context.Context, key string, path string) ([]*int64, error) {
	cmd := resp3.NewRequest(resp3.DataTypeAny, "JSON.STRLEN", key, path)
	resp := c.do(ctx, cmd)
	return parserJSONPtrInt64Slice(resp.result, resp.err)
}

// JSONToggle 将指定 key 中 JSON 路径 path 对应的布尔值取反（true ↔ false）。
//
// path 可以使用两种语法：
//   - 基于 `$` 的 JSONPath（默认）：适用于批量匹配，但 JSONToggleOne 只会取第一个匹配值。
//   - 基于 `.` 的单一路径：直接定位单个值。
//
// 返回值：
//   - 布尔指针 *bool：表示取反后的值（true / false）。
//   - 如果匹配的值不是布尔类型，返回 nil。
//   - 如果路径未匹配到任何值，返回 nil。
func (c *Client) JSONToggle(ctx context.Context, key string, path string) ([]*bool, error) {
	cmd := resp3.NewRequest(resp3.DataTypeAny, "JSON.TOGGLE", key, path)
	resp := c.do(ctx, cmd)
	return parserJSONPtrBoolSlice(resp.result, resp.err)
}

func parserJSONPtrBoolSlice(data resp3.Element, err error) ([]*bool, error) {
	if err != nil {
		return nil, err
	}
	switch rv := data.(type) {
	case resp3.BulkString:
		ptr, err := resp3.ToPtrBool(rv, nil)
		if err != nil {
			return nil, err
		}
		return []*bool{ptr}, nil
	case resp3.Integer:
		ptr, err := resp3.ToPtrBool(rv, nil)
		if err != nil {
			return nil, err
		}
		return []*bool{ptr}, nil
	case resp3.Null:
		// 类型不对
		return []*bool{nil}, nil
	case resp3.Array:
		return resp3.ToPtrBoolSlice(rv, nil, 0)
	case resp3.SimpleError:
		return nil, rv
	default:
		return nil, fmt.Errorf("invalid type for result: %T", rv)
	}
}

// JSONType 返回指定 key 的 JSON 值类型
//
// 参数：
//   - key: 键名
//
// 返回值：
//   - []JSONType: 类型，如 nil( key 或者 path 不存在), string、integer、boolean、object、array
func (c *Client) JSONType(ctx context.Context, key string) ([]any, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "JSON.TYPE", key)
	resp := c.do(ctx, cmd)
	return resp3.ToStringAnySlice(resp.result, resp.err, 0)
}

// JSONTypeWithPath 返回指定 key 下、由 path 定位的 JSON 值类型
//
// 参数：
//   - key: 键名
//   - path: JSON 路径，用于定位 JSON 中的元素。可选，不为空时优先，为空时 Server 使用默认值 "$"。
//
// 返回值：
//   - [][]JSONType: 类型，如 nil( key 或者 path 不存在), string、integer、boolean、object、array
func (c *Client) JSONTypeWithPath(ctx context.Context, key string, path string) ([][]any, error) {
	cmd := resp3.NewRequest(resp3.DataTypeArray, "JSON.TYPE", key, path)
	resp := c.do(ctx, cmd)
	arr, err := resp3.ToSlice(resp.result, resp.err)
	if err != nil || len(arr) == 0 {
		return nil, err
	}
	result := make([][]any, len(arr))
	for i, v := range arr {
		tmp, err := resp3.ToStringAnySlice(v, nil, 0)
		if err != nil {
			return nil, err
		}
		result[i] = tmp
	}
	return result, nil
}
