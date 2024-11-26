//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-01

package xbase62

import (
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/fsgo/fst"
)

func TestEncodeInt64(t *testing.T) {
	numbers := []int64{-1, 0, -999, 999, 9999, 100000, math.MaxInt64}
	for _, num := range numbers {
		t.Run(strconv.FormatInt(num, 10), func(t *testing.T) {
			str := EncodeInt64(num)
			t.Logf("num=%d b62=%q", num, str)
			got, err := DecodeInt64String(str)
			fst.NoError(t, err)
			fst.Equal(t, num, got)
		})
	}
	fst.Equal(t, "7m85Y0n8LzA", EncodeInt64(math.MaxInt64))
}

func TestEncodeToString(t *testing.T) {
	checkEncodeDecode := func(t *testing.T, str string) {
		got1 := EncodeToString([]byte(str))
		got2, err2 := DecodeString(got1)
		fst.NoError(t, err2)
		fst.Equal(t, str, string(got2))

		got2, err2 = Decode([]byte(got1))
		fst.NoError(t, err2)
		fst.Equal(t, str, string(got2))
	}
	checkEncodeDecode(t, "hello 你好")
	checkEncodeDecode(t, "")

	for i := 0; i < 100; i++ {
		str := strings.Repeat("i", i)
		checkEncodeDecode(t, str)
	}
}
