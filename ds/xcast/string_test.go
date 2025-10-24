//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-23

package xcast

import (
	"net"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestString(t *testing.T) {
	testString(t, "abc", "abc", true)

	testString(t, int(123), "123", true)
	testString(t, int8(123), "123", true)
	testString(t, int16(123), "123", true)
	testString(t, int32(123), "123", true)
	testString(t, int64(123), "123", true)

	testString(t, uint(123), "123", true)
	testString(t, uint8(123), "123", true)
	testString(t, uint16(123), "123", true)
	testString(t, uint32(123), "123", true)
	testString(t, uint64(123), "123", true)

	testString(t, 123.1, "123.1", true)
	testString(t, float32(123.1), "123.1", true)

	testString(t, true, "true", true)
	testString(t, false, "false", true)

	testString(t, nil, "", false)
	testString(t, &net.AddrError{}, "", false)

	xt.Equal(t, "abc", ToString("abc"))
}

func testString(t *testing.T, v any, expect string, ok bool) {
	t.Helper()
	got1, got2 := String(v)
	xt.Equal(t, expect, got1)
	xt.Equal(t, ok, got2)
}
