//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-30

// https://github.com/go-sql-driver/mysql/blob/master/conncheck_dummy.go

//go:build !linux && !darwin && !dragonfly && !freebsd && !netbsd && !openbsd && !solaris && !illumos

package znet

import "net"

func ConnCheck(conn net.Conn) error {
	return nil
}
