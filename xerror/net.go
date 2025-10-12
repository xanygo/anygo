//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-12

package xerror

import (
	"errors"
	"io"
	"net"
	"os"
	"syscall"
)

var _ net.Error = (*net.OpError)(nil)
var _ net.Error = (*net.DNSError)(nil)

// IsClientNetError 判断是否是网络错误
func IsClientNetError(err error) bool {
	if err == nil {
		return false
	}
	// 包含了 *net.OpError // 最常见的网络错误
	var ne net.Error
	if errors.As(err, &ne) {
		return true
	}
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return true
	}
	var sysErr *os.SyscallError
	if errors.As(err, &sysErr) {
		return true
	}
	if errors.Is(err, io.EOF) {
		return true
	}
	return false
}
