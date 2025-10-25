//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-25

package zos

import (
	"os"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestIsTerminalFile(t *testing.T) {
	isTerminal(os.Stdout)
	f, err := os.Open("term.go")
	xt.NoError(t, err)
	defer f.Close()
	xt.False(t, IsTerminalFile(f))
}
