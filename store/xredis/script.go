//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-02

package xredis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/store/xredis/resp3"
)

// Eval 调用服务器端 Lua 脚本的执行。
//
// 第一个参数是脚本的源代码。脚本使用 Lua 编写，并由 Redis 内置的 Lua 5.1 解释器执行
func (c *Client) Eval(ctx context.Context, script string, keys []string, args ...any) *Result {
	// EVAL script numkeys [key [key ...]] [arg [arg ...]]
	return c.doEval(ctx, "EVAL", script, keys, args...)
}

func (c *Client) doEval(ctx context.Context, method string, script string, keys []string, args ...any) *Result {
	// EVAL script numkeys [key [key ...]] [arg [arg ...]]
	arr := make([]any, 3, 3+len(keys)+len(args))
	arr[0] = method
	arr[1] = script
	arr[2] = len(keys)
	for _, k := range keys {
		arr = append(arr, k)
	}
	arr = append(arr, args...)
	cmd := resp3.NewRequest(resp3.DataTypeAny, arr...)
	resp := c.do(ctx, cmd)
	return NewResult(resp.result, resp.err)
}

// EvalRO 这是 EVAL 命令的只读变体，它不能执行修改数据的命令
func (c *Client) EvalRO(ctx context.Context, script string, keys []string, args ...any) *Result {
	// EVAL_RO script numkeys [key [key ...]] [arg [arg ...]]
	return c.doEval(ctx, "EVAL_RO", script, keys, args...)
}

// EvalSha 根据脚本的 SHA1 摘要，从服务器的缓存中执行该脚本
func (c *Client) EvalSha(ctx context.Context, sha1 string, keys []string, args ...any) *Result {
	// EVALSHA_RO sha1 numkeys [key [key ...]] [arg [arg ...]]
	return c.doEval(ctx, "EVALSHA", sha1, keys, args...)
}

// FCall 调用函数
//
// 函数通过 FUNCTION LOAD 命令加载到服务器中。第一个参数是已加载函数的名称。
func (c *Client) FCall(ctx context.Context, function string, keys []string, args ...any) *Result {
	// FCALL function numkeys [key [key ...]] [arg [arg ...]]
	return c.doEval(ctx, "FCALL", function, keys, args...)
}

// FCallRO 这是 FCALL 命令的只读变体，它不能执行修改数据的命令
func (c *Client) FCallRO(ctx context.Context, function string, keys []string, args ...any) *Result {
	// FCALL_RO function numkeys [key [key ...]] [arg [arg ...]]
	return c.doEval(ctx, "FCALL_RO", function, keys, args...)
}

func (c *Client) FunctionDelete(ctx context.Context, libraryName string) error {
	// FUNCTION DELETE library-name
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, "FUNCTION", "DELETE", libraryName)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

// FunctionDump 返回已加载函数库的序列化数据(二进制序列化格式，仅能被 FUNCTION RESTORE 识别和使用)。
// 之后可以使用 FUNCTION RESTORE 命令恢复这些序列化的数据。
func (c *Client) FunctionDump(ctx context.Context) (string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, "FUNCTION", "DUMP")
	resp := c.do(ctx, cmd)
	return resp3.ToString(resp.result, resp.err)
}

// FunctionFlush 删除所有库
func (c *Client) FunctionFlush(ctx context.Context, sync bool) error {
	opt := "ASYNC"
	if sync {
		opt = "SYNC"
	}
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, "FUNCTION", "FLUSH", opt)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

// FunctionKill 终止正在执行的函数
//
// 只能用于在执行过程中未修改数据集的函数（因为停止只读函数不会破坏脚本引擎所保证的原子性）。
func (c *Client) FunctionKill(ctx context.Context) error {
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, "FUNCTION", "KILL")
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

// FunctionLoad 将一个函数库加载到 Redis 中。
//
// 该命令有一个必需的参数：实现该函数库的源代码。
// 库的内容必须以 Shebang 语句开头，用于提供库的元数据（例如要使用的引擎类型和库名）。
//
// FUNCTION LOAD "#!lua name=mylib \n redis.register_function('myfunc', function(keys, args) return args[1] end)"
func (c *Client) FunctionLoad(ctx context.Context, replace bool, code string) (string, error) {
	var args []any
	if replace {
		args = []any{"FUNCTION", "LOAD", "REPLACE", code}
	} else {
		args = []any{"FUNCTION", "LOAD", code}
	}
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToString(resp.result, resp.err)
}

// FunctionReStore 从序列化的负载中恢复库
//
// 你可以使用可选的 policy 参数来指定处理已存在库的策略。允许的策略如下：
//
//	APPEND： 将恢复的库追加到现有库中，遇到名称冲突时中止操作。这是默认策略。
//	FLUSH：  在恢复负载之前删除所有现有库。
//	REPLACE：将恢复的库追加到现有库中，如果发生名称冲突则替换已有库。需要注意的是，此策略不会防止函数名冲突，仅针对库名。
func (c *Client) FunctionReStore(ctx context.Context, serializedValue string, policy string) error {
	var args []any
	if policy != "" {
		args = []any{"FUNCTION", "RESTORE", serializedValue, policy}
	} else {
		args = []any{"FUNCTION", "RESTORE", serializedValue}
	}
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

// FunctionStats 返回当前正在运行的函数的信息，以及可用执行引擎的信息
//
// e.g.:
//
//	 1# "running_script" => (nil)
//	 2# "engines" =>
//		  1# "LUA" =>
//		     1# "libraries_count" => (integer) 0
//		     2# "functions_count" => (integer) 0
func (c *Client) FunctionStats(ctx context.Context) (*FunctionStats, error) {
	cmd := resp3.NewRequest(resp3.DataTypeMap, "FUNCTION", "STATS")
	resp := c.do(ctx, cmd)
	mp, err := resp3.ToStringAnyMap(resp.result, resp.err)
	if err != nil {
		return nil, err
	}
	fs := &FunctionStats{}
	if eg, ok := mp["engines"]; ok {
		xmap.Range[string, any](eg, func(k string, v any) bool {
			fse := &FunctionStatsEngine{}
			num := xmap.Range[string, any](v, func(k string, v any) bool {
				switch k {
				case "language":
					fse.Language, _ = v.(string)
				case "functions_count":
					fse.FunctionsCount, _ = v.(int64)
				case "libraries_count":
					fse.LibrariesCount, _ = v.(int64)
				}
				return true
			})
			if num > 0 {
				if fs.Engines == nil {
					fs.Engines = make(map[string]*FunctionStatsEngine, 1)
				}
				fs.Engines[k] = fse
			}
			return true
		})
	}

	if rs, ok := mp["running_script"]; ok {
		rn := &FunctionStatsRunning{}
		num := xmap.Range[string, any](rs, func(key string, val any) bool {
			switch key {
			case "name":
				rn.Name, _ = val.(string)
			case "duration":
				dur, _ := val.(int64)
				rn.Duration = time.Duration(dur) * time.Millisecond
			case "command":
				rn.Command, _ = val.([]string)
			}
			return true
		})
		if num > 0 {
			fs.RunningScript = rn
		}
	}
	return fs, nil
}

type FunctionStats struct {
	Engines       map[string]*FunctionStatsEngine `json:"engines"`
	RunningScript *FunctionStatsRunning           `json:"running_script"`
}

func (fs *FunctionStats) String() string {
	bf, _ := json.Marshal(fs)
	return string(bf)
}

type FunctionStatsEngine struct {
	Language       string `json:"language,omitempty"`
	FunctionsCount int64  `json:"functions_count"`
	LibrariesCount int64  `json:"libraries_count"`
}

type FunctionStatsRunning struct {
	Name     string        `json:"name"`     // the name of the function.
	Command  []string      `json:"command"`  // the command and arguments used for invoking the function.
	Duration time.Duration `json:"duration"` // the function's runtime duration in milliseconds.
}

// ScriptDebug 为之后使用 EVAL 执行的脚本设置调试模式。
//
// Redis 内置了一个完整的 Lua 调试器，代号 LDB，它可以让编写复杂脚本的工作变得更加简单。
// 在调试模式下，Redis 充当远程调试服务器，而客户端（如 redis-cli）可以逐步执行脚本、设置断点、检查变量等
//
// 支持如下模式（mode）：
//
//	YES： 启用 Lua 脚本的非阻塞异步调试（更改不会保存）。
//	SYNC：启用 Lua 脚本的阻塞同步调试（更改会保存到数据中）。
//	NO：  禁用脚本调试模式。
func (c *Client) ScriptDebug(ctx context.Context, mode string) error {
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, "SCRIPT", "DEBUG", mode)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

// ScriptExists 返回脚本缓存中脚本是否存在的信息
func (c *Client) ScriptExists(ctx context.Context, sha1 string) (bool, error) {
	vs, err := c.ScriptExistsMany(ctx, sha1)
	if err != nil {
		return false, err
	}
	return vs[0], nil
}

// ScriptExistsMany 返回脚本缓存中脚本是否存在的信息(同时判断多个)
func (c *Client) ScriptExistsMany(ctx context.Context, sha1 ...string) ([]bool, error) {
	if len(sha1) == 0 {
		return nil, nil
	}
	args := make([]any, 2, 2+len(sha1))
	args[0] = "SCRIPT"
	args[1] = "EXISTS"
	for _, s := range sha1 {
		args = append(args, s)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	nums, err := resp3.ToInt64Slice(resp.result, resp.err, len(sha1))
	if err != nil {
		return nil, err
	}
	if len(nums) != len(sha1) {
		return nil, fmt.Errorf("reply expect has %d, but got %d", len(sha1), len(nums))
	}
	result := make([]bool, len(sha1))
	for i, v := range nums {
		result[i] = v == 1
	}
	return result, nil
}

// ScriptFlush 清空 Lua 脚本缓存
//
// 可以使用以下修饰符之一来明确指定刷新模式：
//
//	ASYNC：异步刷新缓存
//	SYNC：同步刷新缓存
func (c *Client) ScriptFlush(ctx context.Context, sync bool) error {
	opt := "ASYNC"
	if sync {
		opt = "SYNC"
	}
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, "SCRIPT", "FLUSH", opt)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

// ScriptKill 终止当前正在执行的 EVAL 脚本，前提是该脚本尚未执行任何写操作。
//
// 该命令主要用于终止运行时间过长的脚本（例如，由于某个 bug 导致进入了无限循环）。
// 脚本将被终止，而当前被 EVAL 阻塞的客户端会看到该命令返回错误。
func (c *Client) ScriptKill(ctx context.Context) error {
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, "SCRIPT", "KILL")
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

// ScriptLoad 将脚本加载到脚本缓存中，但不执行它。
// 脚本加载到缓存后，可以使用 EVALSHA 结合脚本的正确 SHA1 摘要来调用，就像首次成功调用 EVAL 后一样。
//
//	脚本保证会一直保留在脚本缓存中（除非调用 SCRIPT FLUSH）。
//	即使脚本已经存在于脚本缓存中，该命令的行为也保持不变。
func (c *Client) ScriptLoad(ctx context.Context, script string) (string, error) {
	cmd := resp3.NewRequest(resp3.DataTypeBulkString, "SCRIPT", "LOAD", script)
	resp := c.do(ctx, cmd)
	return resp3.ToString(resp.result, resp.err)
}
