//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-16

package xredis_test

import (
	"context"
	"testing"
	"time"

	"github.com/xanygo/anygo/internal/redistest"
	"github.com/xanygo/anygo/store/xredis"
	"github.com/xanygo/anygo/xt"
)

func TestClient_Script(t *testing.T) {
	ts, errTs := redistest.NewServer()
	if errTs != nil {
		t.Skipf("create redis-server skipped: %v", errTs)
		return
	}
	defer ts.Stop()
	t.Logf("uri= %q", ts.URI())

	_, client, errClient := xredis.NewClientByURI("demo", ts.URI())
	xt.NoError(t, errClient)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	t.Run("EVAL", func(t *testing.T) {
		result := client.Eval(ctx, `return ARGV[1]`, nil, "hello")
		xt.NoError(t, result.Err())
		xt.Equal(t, "hello", result.Value())

		result = client.Eval(ctx, `return ARGV[1]`, nil) // 缺少参数
		xt.NoError(t, result.Err())
		xt.Empty(t, result.Value())
	})

	t.Run("FunctionDump", func(t *testing.T) {
		result, err := client.FunctionDump(ctx)
		xt.NoError(t, err)
		xt.NotEmpty(t, result)
	})

	t.Run("FunctionFlush", func(t *testing.T) {
		err := client.FunctionFlush(ctx, true)
		xt.NoError(t, err)

		err = client.FunctionFlush(ctx, false)
		xt.NoError(t, err)
	})

	t.Run("FunctionKill", func(t *testing.T) {
		err := client.FunctionKill(ctx)
		xt.Error(t, err)
		xt.Contains(t, err.Error(), "NOTBUSY No scripts")
	})

	t.Run("FunctionLoad", func(t *testing.T) {
		code := "#!lua name=mylib \n redis.register_function('myfunc', function(keys, args) return args[1] end)"
		str, err := client.FunctionLoad(ctx, false, code)
		t.Logf("FunctionLoad=%q", str)
		xt.NoError(t, err)
		xt.Equal(t, "mylib", str)

		str, err = client.FunctionLoad(ctx, true, code)
		xt.NoError(t, err)
		xt.Equal(t, "mylib", str)
	})

	t.Run("FunctionStats", func(t *testing.T) {
		st, err := client.FunctionStats(ctx)
		xt.NoError(t, err)
		t.Logf("FunctionStats=%v", st)
	})

	t.Run("ScriptDebug", func(t *testing.T) {
		err := client.ScriptDebug(ctx, "")
		xt.Error(t, err)

		err = client.ScriptDebug(ctx, "yes")
		xt.NoError(t, err)

		err = client.ScriptDebug(ctx, "SYNC")
		xt.NoError(t, err)

		err = client.ScriptDebug(ctx, "NO")
		xt.NoError(t, err)
	})

	t.Run("ScriptExists", func(t *testing.T) {
		val, err := client.ScriptExists(ctx, "hello")
		xt.NoError(t, err)
		xt.False(t, val)

		script := "return 'Hello'"
		sa, err := client.ScriptLoad(ctx, script)
		xt.NoError(t, err)
		xt.NotEmpty(t, sa)

		val, err = client.ScriptExists(ctx, sa)
		t.Logf("ScriptExists(%q)=%v", script, val)
		xt.NoError(t, err)
		xt.True(t, val)
	})

	t.Run("ScriptFlush", func(t *testing.T) {
		err := client.ScriptFlush(ctx, true)
		xt.NoError(t, err)

		err = client.ScriptFlush(ctx, false)
		xt.NoError(t, err)
	})

	t.Run("ScriptKill", func(t *testing.T) {
		err := client.ScriptKill(ctx)
		xt.Error(t, err)
		xt.ErrorContains(t, err, "NOTBUSY")
	})
}
