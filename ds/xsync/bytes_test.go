//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-29

package xsync_test

import (
	"testing"

	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/xt"
)

func TestNewBytesBuffer(t *testing.T) {
	b1 := xsync.NewBytesBufferString("hello")
	xt.Equal(t, "hello", b1.String())
}
