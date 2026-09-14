//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package xcodec

import (
	"bytes"
	"encoding"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestText(t *testing.T) {
	tc := Text
	t.Run("string-1", func(t *testing.T) {
		out, err := tc.Marshal("string")
		xt.NoError(t, err)
		xt.Equal(t, string(out), "string")

		var str string
		err = tc.Unmarshal([]byte("string"), &str)
		xt.NoError(t, err)
		xt.Equal(t, str, "string")
	})

	t.Run("my-string", func(t *testing.T) {
		type myString string
		out, err := tc.Marshal(myString("string"))
		xt.NoError(t, err)
		xt.Equal(t, string(out), "string")

		var str myString
		err = tc.Unmarshal([]byte("string"), &str)
		xt.NoError(t, err)
		xt.Equal(t, str, "string")
	})

	t.Run("int-1", func(t *testing.T) {
		out, err := tc.Marshal(123)
		xt.NoError(t, err)
		xt.Equal(t, string(out), "123")

		var str int
		err = tc.Unmarshal([]byte("123"), &str)
		xt.NoError(t, err)
		xt.Equal(t, str, 123)
	})

	t.Run("bytes", func(t *testing.T) {
		out, err := tc.Marshal([]byte("string"))
		xt.NoError(t, err)
		xt.Equal(t, string(out), "string")

		var str []byte
		err = tc.Unmarshal([]byte("string"), &str)
		xt.NoError(t, err)
		xt.Equal(t, string(str), "string")
	})

	getIntPtr := func(num int64) *int64 {
		return &num
	}
	t.Run("ptr-int-1", func(t *testing.T) {
		itp1 := getIntPtr(123)
		out, err := tc.Marshal(itp1)
		xt.NoError(t, err)
		xt.Equal(t, string(out), "123")

		var num1 *int64
		err = tc.Unmarshal([]byte("123"), &num1)
		xt.NoError(t, err)
		xt.Equal(t, *num1, 123)
	})

	t.Run("myInt1-encode1", func(t *testing.T) {
		out, err := tc.Marshal(MyInt1(123))
		xt.NoError(t, err)
		xt.Equal(t, string(out), "hello-123")
	})
	t.Run("myInt1-encode2", func(t *testing.T) {
		v := MyInt1(123)
		out, err := tc.Marshal(&v)
		xt.NoError(t, err)
		xt.Equal(t, string(out), "hello-123")
	})

	t.Run("myInt1-decode-1", func(t *testing.T) {
		var num1 *MyInt1
		err := tc.Unmarshal([]byte("hello-123"), &num1)
		xt.NoError(t, err)
		xt.Equal(t, *num1, 123)
	})

	t.Run("myInt2-encode1", func(t *testing.T) {
		out, err := tc.Marshal(MyInt2(123))
		xt.NoError(t, err)
		xt.Equal(t, string(out), "world-123")
	})
	t.Run("myInt2-encode2", func(t *testing.T) {
		v := MyInt2(123)
		out, err := tc.Marshal(&v)
		xt.NoError(t, err)
		xt.Equal(t, string(out), "world-123")
	})

	t.Run("myInt2-decode-1", func(t *testing.T) {
		var num1 *MyInt2
		err := tc.Unmarshal([]byte("world-123"), &num1)
		xt.NoError(t, err)
		xt.Equal(t, *num1, 123)
	})

	t.Run("myInt2-decode-2", func(t *testing.T) {
		var num1 MyInt2
		err := tc.Unmarshal([]byte("world-123"), &num1)
		xt.NoError(t, err)
		xt.Equal(t, num1, 123)
	})
}

var _ encoding.TextMarshaler = (*MyInt1)(nil)
var _ encoding.TextUnmarshaler = (*MyInt1)(nil)

type MyInt1 int64

// MyInt 实际是 int64,所以不应该调用 MarshalText 和 UnmarshalText
func (m MyInt1) MarshalText() (text []byte, err error) {
	return fmt.Appendf(nil, "hello-%d", m), nil
}

func (m *MyInt1) UnmarshalText(text []byte) error {
	after, found := bytes.CutPrefix(text, []byte("hello-"))
	if !found {
		return errors.New("miss prefix")
	}
	num, err := strconv.ParseInt(string(after), 10, 64)
	if err != nil {
		return err
	}
	*m = MyInt1(num)
	return nil
}

var _ encoding.TextMarshaler = (*MyInt2)(nil)
var _ encoding.TextUnmarshaler = (*MyInt2)(nil)

type MyInt2 int64

// MyInt 实际是 int64,所以不应该调用 MarshalText 和 UnmarshalText
func (m *MyInt2) MarshalText() (text []byte, err error) {
	return fmt.Appendf(nil, "world-%d", *m), nil
}

func (m *MyInt2) UnmarshalText(text []byte) error {
	after, found := bytes.CutPrefix(text, []byte("world-"))
	if !found {
		return errors.New("miss prefix")
	}
	num, err := strconv.ParseInt(string(after), 10, 64)
	if err != nil {
		return err
	}
	*m = MyInt2(num)
	return nil
}
