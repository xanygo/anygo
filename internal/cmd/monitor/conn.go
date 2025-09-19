//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-18

package monitor

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/xanygo/anygo/xio/xfs"
	"github.com/xanygo/anygo/xnet"
)

type ConnMonitor struct {
	Logger    *log.Logger
	PrintType string
	OutputDir string
	NoRead    bool
	ID        int64
	Time      time.Time
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

	if !c.NoRead {
		c.Logger.Printf(format, readSize, b[:readSize])
	}

	c.writeFile(fmt.Sprintf(format, readSize, b[:readSize]))
}

func (c *ConnMonitor) writeFile(str string) {
	if c.OutputDir == "" {
		return
	}
	xfs.KeepDirExists(c.OutputDir)
	name := fmt.Sprintf("%d_%06d_%s.txt", os.Getpid(), c.ID, c.Time.Format("02150405"))
	filePath := filepath.Join(c.OutputDir, name)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		c.Logger.Println("OpenFile err:", err)
		return
	}
	defer file.Close()
	_, err = file.WriteString(str + "\n\n")
	if err != nil {
		c.Logger.Println("file.WriteString err:", err)
		return
	}
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

	c.writeFile(fmt.Sprintf(format, wroteSize, b[:wroteSize]))
}
