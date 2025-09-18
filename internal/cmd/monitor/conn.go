//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-18

package monitor

import (
	"log"

	"github.com/xanygo/anygo/xnet"
)

type ConnMonitor struct {
	Logger    *log.Logger
	PrintType string //
}

func (c *ConnMonitor) Interceptor() *xnet.ConnInterceptor {
	return &xnet.ConnInterceptor{
		AfterRead:  c.afterRead,
		AfterWrite: c.afterWrite,
	}
}

func (c *ConnMonitor) afterRead(info xnet.ConnInfo, b []byte, readSize int, err error) {
	format := "Read %d Bytes :"
	switch c.PrintType {
	case "b":
		format += "%b"
	case "c":
		format += "%c"
	case "q":
		format += "%q"
	default:
		format += "%s"
	}
	format += "\n\n"
	c.Logger.Printf(format, readSize, b[:readSize])
}

func (c *ConnMonitor) afterWrite(info xnet.ConnInfo, b []byte, wroteSize int, err error) {
	format := "Write %d Bytes :"
	switch c.PrintType {
	case "b":
		format += "%b"
	case "c":
		format += "%c"
	case "q":
		format += "%q"
	default:
		format += "%s"
	}
	format += "\n\n"
	c.Logger.Printf(format, wroteSize, b[:wroteSize])
}
