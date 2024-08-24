//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-24

//go:build !(linux || darwin || windows)

package fctime

import (
	"os"
	"time"
)

func Ctime(st os.FileInfo) time.Time {
	// 其他情况使用文件修改时间
	return st.ModTime()
}
