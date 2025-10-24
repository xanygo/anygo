//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-23

package xcast

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestBool(t *testing.T) {
	testBool(t, "true", true, true)
	testBool(t, "false", false, true)
	testBool(t, true, true, true)
	testBool(t, false, false, true)

	testBool(t, 100, false, false)

	xt.Equal(t, true, ToBool("true"))
}

func testBool(t *testing.T, v any, expect bool, ok bool) {
	t.Helper()
	got1, got2 := Bool(v)
	xt.Equal(t, expect, got1)
	xt.Equal(t, ok, got2)
}
