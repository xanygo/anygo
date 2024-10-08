//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-21

package xlog

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/fsgo/fst"
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
		fst.NoError(t, json.Unmarshal(logContent, &data))
		fst.NotEmpty(t, data["source"])
		fst.Equal(t, "v1", data["k1"])
		fst.Equal(t, "v2", data["k2"])
		fst.Equal(t, "v3", data["k3"])
		fst.Equal(t, "hello", data["msg"])
		fst.Equal(t, level.String(), data["level"].(string))
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
