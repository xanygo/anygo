//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-23

package fctime

import (
	"os"
	"syscall"
	"time"
)

func Ctime(st os.FileInfo) time.Time {
	stat := st.Sys().(*syscall.Win32FileAttributeData)
	return time.Unix(0, stat.CreationTime.Nanoseconds())
}
