//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package xcodec

import (
	"io"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestCSVCodec(t *testing.T) {
	t.Run("string1", func(t *testing.T) {
		str, err := EncodeToString(CSV, []string{"a", "b", "c"})
		xt.NoError(t, err)
		xt.Equal(t, "a,b,c", str)

		var a0 []string
		err = DecodeFromString(CSV, "a,b,c", &a0)
		xt.NoError(t, err)
		xt.Equal(t, []string{"a", "b", "c"}, a0)
	})

	t.Run("int1", func(t *testing.T) {
		var a1 []int
		err := DecodeFromString(CSV, "a,b,c", &a1)
		xt.Error(t, err)
		xt.Empty(t, a1)

		str, err := EncodeToString(CSV, []int{1, 2, 3})
		xt.NoError(t, err)
		xt.Equal(t, "1,2,3", str)

		var b0 []int
		err = DecodeFromString(CSV, "1,2,3", &b0)
		xt.NoError(t, err)
		xt.Equal(t, []int{1, 2, 3}, b0)
	})

	t.Run("int64_1", func(t *testing.T) {
		str, err := EncodeToString(CSV, []int64{1, 2, 3})
		xt.NoError(t, err)
		xt.Equal(t, "1,2,3", str)

		var c0 []int64
		err = DecodeFromString(CSV, "1,2,3", &c0)
		xt.NoError(t, err)
		xt.Equal(t, []int64{1, 2, 3}, c0)
	})

	t.Run("bool_1", func(t *testing.T) {
		str, err := EncodeToString(CSV, []bool{true, false, true})
		xt.NoError(t, err)
		xt.Equal(t, "true,false,true", str)

		var b1 []bool
		err = DecodeFromString(CSV, "true,false,true", &b1)
		xt.NoError(t, err)
		xt.Equal(t, []bool{true, false, true}, b1)
	})

	t.Run("uint8_1", func(t *testing.T) {
		str, err := EncodeToString(CSV, []uint8{1, 2, 3})
		xt.NoError(t, err)
		xt.Equal(t, "1,2,3", str)

		var c1 []uint8
		err = DecodeFromString(CSV, "1,2,3", &c1)
		xt.NoError(t, err)
		xt.Equal(t, []uint8{1, 2, 3}, c1)

		var c2 []uint8
		err = DecodeFromString(CSV, "1024,2,3", &c2)
		xt.Empty(t, c2)
		xt.Error(t, err)
	})

	t.Run("enc-err", func(t *testing.T) {
		str, err := EncodeToString(CSV, []error{io.EOF})
		xt.Empty(t, str)
		xt.Error(t, err)
	})
	t.Run("str-empty", func(t *testing.T) {
		str, err := EncodeToString(CSV, []string{})
		xt.Empty(t, str)
		xt.NoError(t, err)

		var ss []string
		err = DecodeFromString(CSV, "", &ss)
		xt.NoError(t, err)
		xt.Empty(t, ss)
	})
}
