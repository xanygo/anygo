//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-25

//go:build darwin || freebsd || netbsd || openbsd

package zos

import (
	"os"
	"syscall"
	"unsafe"
)

func isTerminal(f *os.File) bool {
	fd := f.Fd()
	var termios syscall.Termios
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TIOCGETA, uintptr(unsafe.Pointer(&termios)))
	return err == 0
}
