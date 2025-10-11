//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-23

package xfs

import (
	"io/fs"
	"time"

	"github.com/xanygo/anygo/internal/fctime"
)

// Ctime 返回文件的创建时间
func Ctime(info fs.FileInfo) time.Time {
	return fctime.Ctime(info)
}
