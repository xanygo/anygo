//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-18

package monitor

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
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
	b = b[:readSize]
	var content []byte
	format := "Read %d Bytes :\n"
	args := []any{readSize}
	for _, pt := range strings.Split(c.PrintType, ",") {
		pt = strings.TrimSpace(pt)
		if pt == "" {
			continue
		}
		content = b
		format += "--->(" + pt + ")\n"
		switch pt {
		case "b":
			format += "%b"
		case "c":
			format += "%c"
		case "q":
			format += "%q"
		case "qn":
			format += "%s"
			s := fmt.Sprintf("%q", content)
			s = strings.ReplaceAll(s, "\\n", "\\n\n")
			content = []byte(s[1 : len(s)-1])
		case "x":
			format += "%x"
		case "X":
			format += "%X"
		case "U":
			format += "%U"
		case "b64", "base64":
			format += "%s"
			b64 := base64.StdEncoding.EncodeToString(b)
			b64 = splitByWidth(b64, 76)
			content = []byte(b64)
		default:
			format += "%s"
		}
		format += "\n<---\n"
		args = append(args, content)
	}

	if !c.NoRead {
		c.Logger.Printf(format, args...)
	}

	c.writeFile(fmt.Sprintf(format, args...))
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
	b = b[:wroteSize]
	var content []byte
	format := "Write %d Bytes:\n"
	args := []any{wroteSize}
	for _, pt := range strings.Split(c.PrintType, ",") {
		pt = strings.TrimSpace(pt)
		if pt == "" {
			continue
		}
		content = b
		format += "--->(" + pt + ")\n"
		switch pt {
		case "b":
			format += "%b"
		case "c":
			format += "%c"
		case "q":
			format += "%q"
		case "qn":
			format += "%s"
			s := fmt.Sprintf("%q", content)
			s = strings.ReplaceAll(s, "\\n", "\\n\n")
			content = []byte(s[1 : len(s)-1])
		case "x":
			format += "%x"
		case "X":
			format += "%X"
		case "U":
			format += "%U"
		case "b64", "base64":
			format += "%s"
			b64 := base64.StdEncoding.EncodeToString(b)
			b64 = splitByWidth(b64, 76)
			content = []byte(b64)
		default:
			format += "%s"
		}
		format += "\n<---\n"
		args = append(args, content)
	}
	c.Logger.Printf(format, args...)
	c.writeFile(fmt.Sprintf(format, args...))
}

func splitByWidth(s string, width int) string {
	var lines []string
	for len(s) > width {
		lines = append(lines, s[:width])
		s = s[width:]
	}
	if len(s) > 0 {
		lines = append(lines, s)
	}
	return strings.Join(lines, "\n")
}
