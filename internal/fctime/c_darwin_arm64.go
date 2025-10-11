//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-23

package fctime

import (
	"os"
	"syscall"
	"time"
)

func Ctime(info os.FileInfo) time.Time {
	stat := info.Sys().(*syscall.Stat_t)
	cs := stat.Ctimespec
	return time.Unix(cs.Sec, cs.Nsec)
}
