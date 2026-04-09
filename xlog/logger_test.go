//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-21

package xlog

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestSimple(t *testing.T) {
	bf := &bytes.Buffer{}
	lg := NewSimple(bf)
	ctx := NewContext(context.Background())
	AddMetaAttr(ctx, String("k1", "v1"))
	AddAttr(ctx, String("k2", "v2"))
	a3 := String("k3", "v3")
	checkLog := func(t *testing.T, level Level) {
		logContent := bf.Bytes()
		data := map[string]any{}
		xt.NoError(t, json.Unmarshal(logContent, &data))
		xt.NotEmpty(t, data["source"])

		meta, ok := data["meta"].(map[string]any)
		xt.True(t, ok)
		xt.Equal(t, meta["k1"], "v1")

		attr, ok := data["attr"].(map[string]any)
		xt.True(t, ok)

		xt.Equal(t, attr["k2"], "v2")
		xt.Equal(t, attr["k3"], "v3")

		xt.Equal(t, data["msg"], "hello")
		xt.Equal(t, data["level"].(string), level.String())
		bf.Reset()
	}
	t.Run("info", func(t *testing.T) {
		lg.Info(ctx, "hello", a3)
		checkLog(t, LevelInfo)
	})

	t.Run("Warn", func(t *testing.T) {
		lg.Warn(ctx, "hello", a3)
		checkLog(t, LevelWarn)
	})

	t.Run("Error", func(t *testing.T) {
		lg.Error(ctx, "hello", a3)
		checkLog(t, LevelError)
	})

	t.Run("Debug", func(t *testing.T) {
		lg.Debug(ctx, "hello", a3)
		checkLog(t, LevelDebug)
	})
}
