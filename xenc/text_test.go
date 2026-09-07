//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-03-17

package xenc_test

import (
	"testing"

	"github.com/xanygo/anygo/xenc"
	"github.com/xanygo/anygo/xenc/xcipher"
	"github.com/xanygo/anygo/xenc/xcodec"
	"github.com/xanygo/anygo/xt"
)

func TestCodecWithCipher(t *testing.T) {
	aes := &xcipher.AesOFB{
		Key: "demo",
	}
	coder := xenc.CodecWithCipher(xcodec.JSON, aes)

	t.Run("case 1", func(t *testing.T) {
		input := "Hello World"
		got1, err := coder.Marshal(input)
		xt.NoError(t, err)
		var str string
		err = coder.Unmarshal(got1, &str)
		xt.NoError(t, err)
		xt.Equal(t, str, input)
	})

	t.Run("case 2", func(t *testing.T) {
		input := map[string]any{
			"a": "hello",
			"b": "你好😄",
		}
		got1, err := coder.Marshal(input)
		xt.NoError(t, err)
		var want map[string]any
		err = coder.Unmarshal(got1, &want)
		xt.NoError(t, err)
		xt.Equal(t, want, input)
	})
	t.Run("case 3", func(t *testing.T) {
		input := map[string]any{
			"a": "hello",
			"b": "你好😄",
		}
		got1, err := xcodec.MarshalToString(coder, input)
		xt.NoError(t, err)
		var want map[string]any
		err = xcodec.UnmarshalFromString(coder, got1, &want)
		xt.NoError(t, err)
		xt.Equal(t, want, input)
	})
}
