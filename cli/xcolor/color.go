//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-25

package xcolor

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/internal/zos"
)

var noColor = os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" || !zos.IsTerminalFile(os.Stdout)

var output atomic.Value

func SetOutput(out io.Writer) {
	output.Store(&out)
}

func Output() io.Writer {
	val := output.Load()
	if val == nil {
		return os.Stderr
	}
	return val.(io.Writer)
}

func DisableColor() {
	noColor = true
}

func New(ids ...Code) *Color {
	c := &Color{
		ids: ids,
	}
	c.init()
	return c
}

type Color struct {
	ids   []Code
	begin []byte
	end   []byte
	endLn []byte
}

const escape = "\x1b"

func (c *Color) init() {
	c.begin = []byte(fmt.Sprintf("%s[%sm", escape, xslice.Join(c.ids, ";")))
	c.end = []byte(fmt.Sprintf("%s[%dm", escape, reset))
	c.endLn = []byte(string(c.end) + "\n")
}

func (c *Color) Add(code ...Code) {
	c.ids = append(c.ids, code...)
	c.init()
}

func (c *Color) Fprintf(w io.Writer, format string, a ...any) (n int, err error) {
	if noColor {
		return fmt.Fprintf(w, format, a...)
	}

	bf := bp.Get()
	defer bp.Put(bf)

	bf.Write(c.begin)
	if len(a) == 0 {
		bf.WriteString(format)
	} else if _, err = fmt.Fprintf(bf, format, a...); err != nil {
		return 0, err
	}
	bf.Write(c.end)

	return w.Write(bf.Bytes())
}

var bp = xsync.NewBytesBufferPool(18 * 1024)

func (c *Color) Fprintln(w io.Writer, a ...any) (n int, err error) {
	if len(a) == 0 {
		return w.Write([]byte("\n"))
	}
	if noColor {
		return fmt.Fprintln(w, a...)
	}
	bf := bp.Get()
	defer bp.Put(bf)

	bf.Write(c.begin)
	if _, err = fmt.Fprint(bf, a...); err != nil {
		return 0, err
	}
	bf.Write(c.endLn)
	return w.Write(bf.Bytes())
}

func (c *Color) Fprint(w io.Writer, a ...any) (n int, err error) {
	if len(a) == 0 {
		return 0, nil
	}
	if noColor {
		return fmt.Fprint(w, a...)
	}

	bf := bp.Get()
	defer bp.Put(bf)

	bf.Write(c.begin)
	if _, err = fmt.Fprint(bf, a...); err != nil {
		return 0, err
	}
	bf.Write(c.end)
	return w.Write(bf.Bytes())
}

func (c *Color) Printf(format string, a ...any) (n int, err error) {
	out := Output()
	return c.Fprintf(out, format, a...)
}

func (c *Color) Println(a ...any) (n int, err error) {
	out := Output()
	return c.Fprintln(out, a...)
}

func (c *Color) Sprintf(format string, a ...any) string {
	bf := bp.Get()
	defer bp.Put(bf)
	c.Fprintf(bf, format, a...)
	return bf.String()
}

func (c *Color) Sprint(a ...any) string {
	bf := bp.Get()
	defer bp.Put(bf)
	c.Fprint(bf, a...)
	return bf.String()
}

var cache = &sync.Map{}

func getByID(id Code) *Color {
	v, ok := cache.Load(id)
	if ok {
		return v.(*Color)
	}
	c := New(id)
	cache.Store(id, c)
	return c
}

func printfln(c *Color, format string, a ...any) {
	c.Println(fmt.Sprintf(format, a...))
}
