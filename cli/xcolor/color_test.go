//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-25

package xcolor

import (
	"fmt"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestBlack(t *testing.T) {
	str := fmt.Sprintf("hello %s", "world")
	Red(str)
	xt.NotEmpty(t, RedString(str))
	HiRed(str)
	BgHiRed(str)
}
