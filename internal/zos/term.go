//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-25

package zos

import "os"

// IsTerminalFile 判断 *os.File 是否是终端
func IsTerminalFile(f *os.File) bool {
	return isTerminal(f)
}
