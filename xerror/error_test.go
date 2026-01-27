//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-20

package xerror

import (
	"io"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestIsTemporary(t *testing.T) {
	xt.False(t, IsTemporary(nil))
	xt.False(t, IsTemporary(io.EOF))

	xt.True(t, IsTemporary(WithTemporary(io.EOF, true)))
	xt.False(t, IsTemporary(WithTemporary(io.EOF, false)))
}
