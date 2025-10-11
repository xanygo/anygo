//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-02

package xredis

import (
	"context"
	"fmt"

	"github.com/xanygo/anygo/store/xredis/resp3"
)

// Eval 调用服务器端 Lua 脚本的执行。
//
// 第一个参数是脚本的源代码。脚本使用 Lua 编写，并由 Redis 内置的 Lua 5.1 解释器执行
func (c *Client) Eval(ctx context.Context, script string, keys []string, args ...any) (Result, error) {
	// EVAL script numkeys [key [key ...]] [arg [arg ...]]
	return c.doEval(ctx, "EVAL", script, keys, args...)
}

func (c *Client) doEval(ctx context.Context, method string, script string, keys []string, args ...any) (Result, error) {
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
	return resp.result, resp.err
}

// EvalRO 这是 EVAL 命令的只读变体，它不能执行修改数据的命令
func (c *Client) EvalRO(ctx context.Context, script string, keys []string, args ...any) (Result, error) {
	// EVAL_RO script numkeys [key [key ...]] [arg [arg ...]]
	return c.doEval(ctx, "EVAL_RO", script, keys, args...)
}

// EvalSha 根据脚本的 SHA1 摘要，从服务器的缓存中执行该脚本
func (c *Client) EvalSha(ctx context.Context, sha1 string, keys []string, args ...any) (Result, error) {
	// EVALSHA_RO sha1 numkeys [key [key ...]] [arg [arg ...]]
	return c.doEval(ctx, "EVALSHA", sha1, keys, args...)
}

// FCall 调用函数
//
// 函数通过 FUNCTION LOAD 命令加载到服务器中。第一个参数是已加载函数的名称。
func (c *Client) FCall(ctx context.Context, function string, keys []string, args ...any) (Result, error) {
	// FCALL function numkeys [key [key ...]] [arg [arg ...]]
	return c.doEval(ctx, "FCALL", function, keys, args...)
}

// FCallRO 这是 FCALL 命令的只读变体，它不能执行修改数据的命令
func (c *Client) FCallRO(ctx context.Context, function string, keys []string, args ...any) (Result, error) {
	// FCALL_RO function numkeys [key [key ...]] [arg [arg ...]]
	return c.doEval(ctx, "FCALL_RO", function, keys, args...)
}

func (c *Client) FunctionDelete(ctx context.Context, libraryName string) error {
	// FUNCTION DELETE library-name
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, "FUNCTION", "DELETE", libraryName)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

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
// APPEND：将恢复的库追加到现有库中，遇到名称冲突时中止操作。这是默认策略。
// FLUSH：在恢复负载之前删除所有现有库。
// REPLACE：将恢复的库追加到现有库中，如果发生名称冲突则替换已有库。需要注意的是，此策略不会防止函数名冲突，仅针对库名。
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
func (c *Client) FunctionStats(ctx context.Context) (resp3.Map, error) {
	cmd := resp3.NewRequest(resp3.DataTypeMap, "FUNCTION", "STATS")
	resp := c.do(ctx, cmd)
	if resp.err != nil {
		return nil, resp.err
	}
	mp, ok := resp.result.(resp3.Map)
	if !ok {
		return mp, nil
	}
	return nil, fmt.Errorf("invalid reply %#v", resp.result)
}

// ScriptDebug 为之后使用 EVAL 执行的脚本设置调试模式。
// Redis 内置了一个完整的 Lua 调试器，代号 LDB，它可以让编写复杂脚本的工作变得更加简单。
// 在调试模式下，Redis 充当远程调试服务器，而客户端（如 redis-cli）可以逐步执行脚本、设置断点、检查变量等
//
//	 支持如下模式（mode）：
//		  YES： 启用 Lua 脚本的非阻塞异步调试（更改不会保存）。
//		  SYNC：启用 Lua 脚本的阻塞同步调试（更改会保存到数据中）。
//		  NO：  禁用脚本调试模式。
func (c *Client) ScriptDebug(ctx context.Context, mode string) error {
	cmd := resp3.NewRequest(resp3.DataTypeSimpleString, "SCRIPT", "DEBUG", mode)
	resp := c.do(ctx, cmd)
	return resp3.ToOkStatus(resp.result, resp.err)
}

func (c *Client) ScriptExists(ctx context.Context, sha1 ...string) ([]int64, error) {
	args := make([]any, 2, 2+len(sha1))
	args[0] = "SCRIPT"
	args[1] = "EXISTS"
	for _, s := range sha1 {
		args = append(args, s)
	}
	cmd := resp3.NewRequest(resp3.DataTypeArray, args...)
	resp := c.do(ctx, cmd)
	return resp3.ToInt64Slice(resp.result, resp.err)
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
