//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package xcodec

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestText(t *testing.T) {
	tc := Text
	t.Run("string-1", func(t *testing.T) {
		out, err := tc.Encode("string")
		xt.NoError(t, err)
		xt.Equal(t, "string", string(out))

		var str string
		err = tc.Decode([]byte("string"), &str)
		xt.NoError(t, err)
		xt.Equal(t, "string", str)
	})

	t.Run("my-string", func(t *testing.T) {
		type myString string
		out, err := tc.Encode(myString("string"))
		xt.NoError(t, err)
		xt.Equal(t, "string", string(out))

		var str myString
		err = tc.Decode([]byte("string"), &str)
		xt.NoError(t, err)
		xt.Equal(t, "string", str)
	})

	t.Run("int-1", func(t *testing.T) {
		out, err := tc.Encode(123)
		xt.NoError(t, err)
		xt.Equal(t, "123", string(out))

		var str int
		err = tc.Decode([]byte("123"), &str)
		xt.NoError(t, err)
		xt.Equal(t, 123, str)
	})

	t.Run("bytes", func(t *testing.T) {
		out, err := tc.Encode([]byte("string"))
		xt.NoError(t, err)
		xt.Equal(t, "string", string(out))

		var str []byte
		err = tc.Decode([]byte("string"), &str)
		xt.NoError(t, err)
		xt.Equal(t, "string", string(str))
	})

	getIntPtr := func(num int64) *int64 {
		return &num
	}
	t.Run("ptr-int-1", func(t *testing.T) {
		itp1 := getIntPtr(123)
		out, err := tc.Encode(itp1)
		xt.NoError(t, err)
		xt.Equal(t, "123", string(out))

		var num1 *int64
		err = tc.Decode([]byte("123"), &num1)
		xt.NoError(t, err)
		xt.Equal(t, 123, *num1)
	})
}
