//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-11

package xcodec

import (
	"net/url"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestFormCodec_Encode(t *testing.T) {
	got1, err1 := Form.Encode(map[string]string{"a": "a", "b": "b"})
	xt.NoError(t, err1)
	xt.Equal(t, "a=a&b=b", string(got1))

	got2, err2 := Form.Encode(url.Values{"a": []string{"a"}, "b": []string{"b"}})
	xt.NoError(t, err2)
	xt.Equal(t, "a=a&b=b", string(got2))

	got3, err3 := Form.Encode("abc")
	xt.Error(t, err3)
	xt.Empty(t, got3)
}

func TestFormCodec_Decode(t *testing.T) {
	var got1 url.Values
	err1 := Form.Decode([]byte("a=a"), &got1)
	xt.NoError(t, err1)
	xt.Equal(t, url.Values{"a": []string{"a"}}, got1)

	var got2 map[string]string
	err2 := Form.Decode([]byte("a=a"), &got2)
	xt.NoError(t, err2)
	xt.Equal(t, map[string]string{"a": "a"}, got2)

	var got3 map[string]any
	err3 := Form.Decode([]byte("a=a"), &got3)
	xt.Error(t, err3)
	xt.Empty(t, got3)
}

func Test_Raw(t *testing.T) {
	str := "hello"
	got1, err1 := Raw.Encode(str)
	xt.NoError(t, err1)
	xt.Equal(t, "hello", string(got1))
	var got2 []byte
	err2 := Raw.Decode(got1, &got2)
	xt.NoError(t, err2)
	xt.Equal(t, "hello", string(got2))
}

func TestEncodeToString(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		got1, err1 := EncodeToString(JSON, "hello")
		xt.NoError(t, err1)
		xt.Equal(t, "hello", got1)

		var got2 string
		err2 := DecodeFromString(JSON, "hello", &got2)
		xt.NoError(t, err2)
		xt.Equal(t, "hello", got2)
	})

	t.Run("bytes", func(t *testing.T) {
		got3, err3 := EncodeToString(JSON, []byte("hello"))
		xt.NoError(t, err3)
		xt.Equal(t, `"aGVsbG8="`, got3)

		var db []byte
		err4 := DecodeFromString(JSON, `"aGVsbG8="`, &db)
		xt.NoError(t, err4)
		xt.Equal(t, "hello", string(db))
	})

	t.Run("my-string", func(t *testing.T) {
		type myString string
		got5, err5 := EncodeToString(JSON, myString("hello"))
		xt.NoError(t, err5)
		xt.Equal(t, "hello", got5)

		var s1 myString
		err6 := DecodeFromString(JSON, `hello`, &s1)
		xt.NoError(t, err6)
		xt.Equal(t, "hello", string(s1))
	})
	t.Run("my-string-ptr", func(t *testing.T) {
		type myString string
		str := myString("hello")
		got5, err5 := EncodeToString(JSON, &str)
		xt.NoError(t, err5)
		xt.Equal(t, "hello", got5)

		var s1 myString
		err6 := DecodeFromString(JSON, `hello`, &s1)
		xt.NoError(t, err6)
		xt.Equal(t, "hello", string(s1))
	})
}
