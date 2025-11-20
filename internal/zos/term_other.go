//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-20

//go:build !(darwin || freebsd || netbsd || openbsd || windows || linux)

package zos

import "os"

func isTerminal(f *os.File) bool {
	return false
}
