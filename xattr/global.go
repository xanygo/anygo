//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-30

package xattr

import (
	"math/rand/v2"
	"time"
)

var rid int64
var startTime time.Time

func init() {
	startTime = time.Now().Local()
	rid = rand.Int64()
}

// RandID 随机数，每次进程重启后发生变化
func RandID() int64 {
	return rid
}

// StartTime 进程启动时间
func StartTime() time.Time {
	return startTime
}
