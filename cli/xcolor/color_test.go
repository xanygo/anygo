//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-25

package xcolor_test

import (
	"fmt"
	"testing"

	"github.com/xanygo/anygo/cli/xcolor"
)

func TestBlack(t *testing.T) {
	str := fmt.Sprintf("hello %s", "world")
	xcolor.Red(str)
	str2 := xcolor.RedString(str)
	if str2 == "" {
		t.Logf("unexpect empty string")
	}
	xcolor.HiRed(str)
	xcolor.BgHiRed(str)
}
