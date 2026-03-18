//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-03-17

package xcodec_test

import (
	"testing"

	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xt"
)

func TestCodecWithCipher(t *testing.T) {
	aes := &xcodec.AesOFB{
		Key: "demo",
	}
	coder := xcodec.CodecWithCipher(xcodec.JSON, aes)

	t.Run("case 1", func(t *testing.T) {
		input := "Hello World"
		got1, err := coder.Encode(input)
		xt.NoError(t, err)
		var str string
		err = coder.Decode(got1, &str)
		xt.NoError(t, err)
		xt.Equal(t, input, str)
	})

	t.Run("case 2", func(t *testing.T) {
		input := map[string]any{
			"a": "hello",
			"b": "你好😄",
		}
		got1, err := coder.Encode(input)
		xt.NoError(t, err)
		var want map[string]any
		err = coder.Decode(got1, &want)
		xt.NoError(t, err)
		xt.Equal(t, input, want)
	})
	t.Run("case 3", func(t *testing.T) {
		input := map[string]any{
			"a": "hello",
			"b": "你好😄",
		}
		got1, err := xcodec.EncodeToString(coder, input)
		xt.NoError(t, err)
		var want map[string]any
		err = xcodec.DecodeFromString(coder, got1, &want)
		xt.NoError(t, err)
		xt.Equal(t, input, want)
	})
}
