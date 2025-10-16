//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-16

package xredis_test

import (
	"context"
	"testing"
	"time"

	"github.com/fsgo/fst"

	"github.com/xanygo/anygo/internal/redistest"
	"github.com/xanygo/anygo/store/xredis"
)

func TestClient_Script(t *testing.T) {
	ts, errTs := redistest.NewServer()
	if errTs != nil {
		t.Logf("create redis fail: %v", errTs)
		return
	}
	defer ts.Stop()
	t.Logf("uri= %q", ts.URI())

	_, client, errClient := xredis.NewClientByURI("demo", ts.URI())
	fst.NoError(t, errClient)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	t.Run("EVAL", func(t *testing.T) {
		result := client.Eval(ctx, `return ARGV[1]`, nil, "hello")
		fst.NoError(t, result.Err())
		fst.Equal(t, "hello", result.Value())

		result = client.Eval(ctx, `return ARGV[1]`, nil) // 缺少参数
		fst.Error(t, result.Err())
		fst.Empty(t, result.Value())
	})

	t.Run("FunctionDump", func(t *testing.T) {
		result, err := client.FunctionDump(ctx)
		fst.NoError(t, err)
		fst.NotEmpty(t, result)
	})

	t.Run("FunctionFlush", func(t *testing.T) {
		err := client.FunctionFlush(ctx, true)
		fst.NoError(t, err)

		err = client.FunctionFlush(ctx, false)
		fst.NoError(t, err)
	})

	t.Run("FunctionKill", func(t *testing.T) {
		err := client.FunctionKill(ctx)
		fst.Error(t, err)
		fst.Contains(t, err.Error(), "NOTBUSY No scripts")
	})

	t.Run("FunctionLoad", func(t *testing.T) {
		code := "#!lua name=mylib \n redis.register_function('myfunc', function(keys, args) return args[1] end)"
		str, err := client.FunctionLoad(ctx, false, code)
		t.Logf("FunctionLoad=%q", str)
		fst.NoError(t, err)
		fst.Equal(t, "mylib", str)

		str, err = client.FunctionLoad(ctx, true, code)
		fst.NoError(t, err)
		fst.Equal(t, "mylib", str)
	})

	t.Run("FunctionStats", func(t *testing.T) {
		st, err := client.FunctionStats(ctx)
		fst.NoError(t, err)
		t.Logf("FunctionStats=%v", st)
	})

	t.Run("ScriptDebug", func(t *testing.T) {
		err := client.ScriptDebug(ctx, "")
		fst.Error(t, err)

		err = client.ScriptDebug(ctx, "yes")
		fst.NoError(t, err)

		err = client.ScriptDebug(ctx, "SYNC")
		fst.NoError(t, err)

		err = client.ScriptDebug(ctx, "NO")
		fst.NoError(t, err)
	})

	t.Run("ScriptExists", func(t *testing.T) {
		val, err := client.ScriptExists(ctx, "hello")
		fst.NoError(t, err)
		fst.False(t, val)

		script := "return 'Hello'"
		sa, err := client.ScriptLoad(ctx, script)
		fst.NoError(t, err)
		fst.NotEmpty(t, sa)

		val, err = client.ScriptExists(ctx, sa)
		t.Logf("ScriptExists(%q)=%v", script, val)
		fst.NoError(t, err)
		fst.True(t, val)
	})

	t.Run("ScriptFlush", func(t *testing.T) {
		err := client.ScriptFlush(ctx, true)
		fst.NoError(t, err)

		err = client.ScriptFlush(ctx, false)
		fst.NoError(t, err)
	})

	t.Run("ScriptKill", func(t *testing.T) {
		err := client.ScriptKill(ctx)
		fst.Error(t, err)
		fst.ErrorContains(t, err, "NOTBUSY")
	})
}
